package repository

import (
	"context"
	"fmt"

	"traffic-backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TrafficRepository struct {
	db *pgxpool.Pool
}

func NewTrafficRepository(db *pgxpool.Pool) *TrafficRepository {
	return &TrafficRepository{
		db: db,
	}
}

func (r *TrafficRepository) SaveTrafficEvent(
	ctx context.Context,
	event model.TrafficEvent,
	locationBucket string,
	timeBucket string,
	vehicleCount int,
	congestionScore int,
) (bool, error) {

	result, err := r.db.Exec(ctx, `
		INSERT INTO traffic_observations (
			event_id,
			bus_id,
			event_timestamp,
			latitude,
			longitude,
			location_bucket,
			time_bucket,
			motorcycle_count,
			car_count,
			truck_count,
			bus_count,
			vehicle_count,
			congestion_score,
			occupancy,
			video_id
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15
		)
		ON CONFLICT (event_id) DO NOTHING
	`,
		event.EventID,
		event.BusID,
		event.Timestamp,
		event.Location.Latitude,
		event.Location.Longitude,
		locationBucket,
		timeBucket,
		event.Vehicles.Motorcycle,
		event.Vehicles.Car,
		event.Vehicles.Truck,
		event.Vehicles.Bus,
		vehicleCount,
		congestionScore,
		event.Traffic.VehicleOccupancy,
		event.Source.VideoID,
	)

	if err != nil {
		return false, err
	}

	return result.RowsAffected() == 1, nil
}

func (r *TrafficRepository) UpdateTrafficAggregate(
	ctx context.Context,
	locationBucket string,
	timeBucket string,
	trafficDate string,
	latitude float64,
	longitude float64,
	vehicleCount int,
	congestionScore int,
	occupancy float64,
) error {

	_, err := r.db.Exec(ctx, `
		INSERT INTO traffic_aggregates (
			location_bucket,
			time_bucket,
			traffic_date,
			latitude,
			longitude,
			observation_count,
			sum_vehicle_count,
			average_vehicle_count,
			sum_congestion_score,
			average_congestion_score,
			sum_occupancy,
			average_occupancy
		)
		VALUES (
			$1, $2, $3,
			$4, $5,
			1,
			$6::BIGINT,
			$6::DOUBLE PRECISION,
			$7::BIGINT,
			$7::DOUBLE PRECISION,
			$8::DOUBLE PRECISION,
			$8::DOUBLE PRECISION
		)
		ON CONFLICT (location_bucket, time_bucket, traffic_date)
		DO UPDATE SET
			observation_count =
				traffic_aggregates.observation_count + 1,

			sum_vehicle_count =
				traffic_aggregates.sum_vehicle_count
				+ EXCLUDED.sum_vehicle_count,

			average_vehicle_count =
				(
					traffic_aggregates.sum_vehicle_count
					+ EXCLUDED.sum_vehicle_count
				)::DOUBLE PRECISION
				/
				(
					traffic_aggregates.observation_count + 1
				),

			sum_congestion_score =
				traffic_aggregates.sum_congestion_score
				+ EXCLUDED.sum_congestion_score,

			average_congestion_score =
				(
					traffic_aggregates.sum_congestion_score
					+ EXCLUDED.sum_congestion_score
				)::DOUBLE PRECISION
				/
				(
					traffic_aggregates.observation_count + 1
				),

			sum_occupancy =
				traffic_aggregates.sum_occupancy
				+ EXCLUDED.sum_occupancy,

			average_occupancy =
				(
					traffic_aggregates.sum_occupancy
					+ EXCLUDED.sum_occupancy
				)
				/
				(
					traffic_aggregates.observation_count + 1
				),

			updated_at = NOW()
	`,
		locationBucket,
		timeBucket,
		trafficDate,
		latitude,
		longitude,
		vehicleCount,
		congestionScore,
		occupancy,
	)

	return err
}
func (r *TrafficRepository) GetHistoricalAverage(
	ctx context.Context,
	locationBucket string,
	timeBucket string,
	currentDate string,
) (float64, float64, error) {

	var averageVehicleCount float64
	var averageCongestionScore float64

	err := r.db.QueryRow(ctx, `
		SELECT
			COALESCE(AVG(average_vehicle_count), 0),
			COALESCE(AVG(average_congestion_score), 0)
		FROM traffic_aggregates
		WHERE location_bucket = $1
		  AND time_bucket = $2
		  AND traffic_date >= ($3::DATE - INTERVAL '3 days')
		  AND traffic_date < $3::DATE
	`,
		locationBucket,
		timeBucket,
		currentDate,
	).Scan(
		&averageVehicleCount,
		&averageCongestionScore,
	)

	if err != nil {
		return 0, 0, err
	}

	return averageVehicleCount, averageCongestionScore, nil
}

func (r *TrafficRepository) GetCurrentTraffic(
	ctx context.Context,
) ([]model.TrafficAggregate, error) {

	rows, err := r.db.Query(ctx, `
		SELECT
			location_bucket,
			latitude,
			longitude,
			time_bucket,
			traffic_date,
			average_vehicle_count,
			average_congestion_score,
			average_occupancy,
			updated_at
		FROM traffic_aggregates
		WHERE traffic_date = CURRENT_DATE
		ORDER BY updated_at DESC
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.TrafficAggregate

	for rows.Next() {
		var aggregate model.TrafficAggregate

		err := rows.Scan(
			&aggregate.LocationBucket,
			&aggregate.Latitude,
			&aggregate.Longitude,
			&aggregate.TimeBucket,
			&aggregate.TrafficDate,
			&aggregate.AverageVehicleCount,
			&aggregate.AverageCongestionScore,
			&aggregate.AverageOccupancy,
			&aggregate.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		results = append(results, aggregate)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *TrafficRepository) SaveEventAndUpdateAggregate(
	ctx context.Context,
	event model.TrafficEvent,
	locationBucket string,
	timeBucket string,
	trafficDate string,
	vehicleCount int,
	congestionScore int,
	occupancy float64,
) error {

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Save raw event
	result, err := tx.Exec(ctx, `
		INSERT INTO traffic_observations (
			event_id,
			bus_id,
			event_timestamp,
			latitude,
			longitude,
			location_bucket,
			time_bucket,
			motorcycle_count,
			car_count,
			truck_count,
			bus_count,
			vehicle_count,
			congestion_score,
			occupancy,
			video_id
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15
		)
		ON CONFLICT (event_id) DO NOTHING
	`,
		event.EventID,
		event.BusID,
		event.Timestamp,
		event.Location.Latitude,
		event.Location.Longitude,
		locationBucket,
		timeBucket,
		event.Vehicles.Motorcycle,
		event.Vehicles.Car,
		event.Vehicles.Truck,
		event.Vehicles.Bus,
		vehicleCount,
		congestionScore,
		occupancy,
		event.Source.VideoID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("duplicate event_id")
	}

	// 2. Update aggregate

	// 2. Update aggregate
	_, err = tx.Exec(ctx, `
		INSERT INTO traffic_aggregates (
			location_bucket,
			time_bucket,
			traffic_date,
			latitude,
			longitude,
			observation_count,
			sum_vehicle_count,
			average_vehicle_count,
			sum_congestion_score,
			average_congestion_score,
			sum_occupancy,
			average_occupancy
		)
		VALUES (
			$1,$2,$3,$4,$5,
			1,
			$6::BIGINT,
			$6::DOUBLE PRECISION,
			$7::BIGINT,
			$7::DOUBLE PRECISION,
			$8::DOUBLE PRECISION,
			$8::DOUBLE PRECISION
		)
		ON CONFLICT (location_bucket, time_bucket, traffic_date)
		DO UPDATE SET

			observation_count =
				traffic_aggregates.observation_count + 1,

			sum_vehicle_count =
				traffic_aggregates.sum_vehicle_count
				+ EXCLUDED.sum_vehicle_count,

			average_vehicle_count =
				(
					traffic_aggregates.sum_vehicle_count
					+ EXCLUDED.sum_vehicle_count
				)::DOUBLE PRECISION
				/
				(traffic_aggregates.observation_count + 1),

			sum_congestion_score =
				traffic_aggregates.sum_congestion_score
				+ EXCLUDED.sum_congestion_score,

			average_congestion_score =
				(
					traffic_aggregates.sum_congestion_score
					+ EXCLUDED.sum_congestion_score
				)::DOUBLE PRECISION
				/
				(traffic_aggregates.observation_count + 1),

			sum_occupancy =
				traffic_aggregates.sum_occupancy
				+ EXCLUDED.sum_occupancy,

			average_occupancy =
				(
					traffic_aggregates.sum_occupancy
					+ EXCLUDED.sum_occupancy
				)
				/
				(traffic_aggregates.observation_count + 1),

			updated_at = NOW()
	`,
		locationBucket,
		timeBucket,
		trafficDate,
		event.Location.Latitude,
		event.Location.Longitude,
		vehicleCount,
		congestionScore,
		occupancy,
	)
	if err != nil {
		return err
	}

	// 3. Commit both operations together
	return tx.Commit(ctx)
}

func (r *TrafficRepository) GetTrafficAggregate(
	ctx context.Context,
	locationBucket string,
	timeBucket string,
	trafficDate string,
) (model.TrafficAggregate, error) {

	var aggregate model.TrafficAggregate

	err := r.db.QueryRow(ctx, `
		SELECT
			location_bucket,
			latitude,
			longitude,
			time_bucket,
			traffic_date,
			average_vehicle_count,
			average_congestion_score,
			average_occupancy,
			updated_at
		FROM traffic_aggregates
		WHERE location_bucket = $1
		  AND time_bucket = $2
		  AND traffic_date = $3::DATE
	`,
		locationBucket,
		timeBucket,
		trafficDate,
	).Scan(
		&aggregate.LocationBucket,
		&aggregate.Latitude,
		&aggregate.Longitude,
		&aggregate.TimeBucket,
		&aggregate.TrafficDate,
		&aggregate.AverageVehicleCount,
		&aggregate.AverageCongestionScore,
		&aggregate.AverageOccupancy,
		&aggregate.UpdatedAt,
	)

	if err != nil {
		return model.TrafficAggregate{}, err
	}

	return aggregate, nil
}

func (r *TrafficRepository) DeleteOldTrafficData(ctx context.Context) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM traffic_observations
		WHERE event_timestamp < NOW() - INTERVAL '3 days'
	`)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, `
		DELETE FROM traffic_aggregates
		WHERE traffic_date < CURRENT_DATE - INTERVAL '3 days'
	`)
	if err != nil {
		return err
	}

	return nil
}
