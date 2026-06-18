package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"crossbow-simulation/backend/internal/model"
)

func (r *Repository) ListCrossbowVariants() ([]model.VariantRecord, error) {
	query := `
		SELECT code, name, dynasty, era_year, description,
			   draw_weight_n, max_range_m, effective_range_m,
			   ideal_fire_rate, magazine_size, reload_time_sec,
			   accuracy_score, mechanism_params, created_at
		FROM crossbow_variants
		ORDER BY created_at ASC
	`

	rows, err := r.db.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to list crossbow variants: %w", err)
	}
	defer rows.Close()

	var variants []model.VariantRecord
	for rows.Next() {
		var v model.VariantRecord
		var dynasty, description *string
		var eraYear *int
		var drawWeightN, maxRangeM, effectiveRangeM *float64
		var idealFireRate, reloadTimeSec, accuracyScore *float64
		var magazineSize *int
		var mechanismParams []byte

		err := rows.Scan(
			&v.Code,
			&v.Name,
			&dynasty,
			&eraYear,
			&description,
			&drawWeightN,
			&maxRangeM,
			&effectiveRangeM,
			&idealFireRate,
			&magazineSize,
			&reloadTimeSec,
			&accuracyScore,
			&mechanismParams,
			&v.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan variant row: %w", err)
		}

		if dynasty != nil {
			v.Dynasty = *dynasty
		}
		if eraYear != nil {
			v.EraYear = *eraYear
		}
		if description != nil {
			v.Description = *description
		}
		if drawWeightN != nil {
			v.DrawWeightN = *drawWeightN
		}
		if maxRangeM != nil {
			v.MaxRangeM = *maxRangeM
		}
		if effectiveRangeM != nil {
			v.EffectiveRangeM = *effectiveRangeM
		}
		if idealFireRate != nil {
			v.IdealFireRate = *idealFireRate
		}
		if magazineSize != nil {
			v.MagazineSize = *magazineSize
		}
		if reloadTimeSec != nil {
			v.ReloadTimeSec = *reloadTimeSec
		}
		if accuracyScore != nil {
			v.AccuracyScore = *accuracyScore
		}
		if len(mechanismParams) > 0 {
			v.MechanismParams = mechanismParams
		}

		variants = append(variants, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return variants, nil
}

func (r *Repository) GetCrossbowVariant(code string) (*model.VariantRecord, error) {
	if code == "" {
		return nil, errors.New("variant code is required")
	}

	query := `
		SELECT code, name, dynasty, era_year, description,
			   draw_weight_n, max_range_m, effective_range_m,
			   ideal_fire_rate, magazine_size, reload_time_sec,
			   accuracy_score, mechanism_params, created_at
		FROM crossbow_variants
		WHERE code = $1
	`

	var v model.VariantRecord
	var dynasty, description *string
	var eraYear *int
	var drawWeightN, maxRangeM, effectiveRangeM *float64
	var idealFireRate, reloadTimeSec, accuracyScore *float64
	var magazineSize *int
	var mechanismParams []byte

	err := r.db.Pool.QueryRow(context.Background(), query, code).Scan(
		&v.Code,
		&v.Name,
		&dynasty,
		&eraYear,
		&description,
		&drawWeightN,
		&maxRangeM,
		&effectiveRangeM,
		&idealFireRate,
		&magazineSize,
		&reloadTimeSec,
		&accuracyScore,
		&mechanismParams,
		&v.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("variant not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get variant: %w", err)
	}

	if dynasty != nil {
		v.Dynasty = *dynasty
	}
	if eraYear != nil {
		v.EraYear = *eraYear
	}
	if description != nil {
		v.Description = *description
	}
	if drawWeightN != nil {
		v.DrawWeightN = *drawWeightN
	}
	if maxRangeM != nil {
		v.MaxRangeM = *maxRangeM
	}
	if effectiveRangeM != nil {
		v.EffectiveRangeM = *effectiveRangeM
	}
	if idealFireRate != nil {
		v.IdealFireRate = *idealFireRate
	}
	if magazineSize != nil {
		v.MagazineSize = *magazineSize
	}
	if reloadTimeSec != nil {
		v.ReloadTimeSec = *reloadTimeSec
	}
	if accuracyScore != nil {
		v.AccuracyScore = *accuracyScore
	}
	if len(mechanismParams) > 0 {
		v.MechanismParams = mechanismParams
	}

	return &v, nil
}

func (r *Repository) SaveMagazineReliabilityAnalysis(analysis model.ReliabilityAnalysisRecord) (string, error) {
	if analysis.CrossbowVariantCode == "" {
		return "", errors.New("crossbow variant code is required")
	}
	if analysis.SimShots <= 0 {
		return "", errors.New("sim_shots must be positive")
	}
	if analysis.SimTimeSec <= 0 {
		return "", errors.New("sim_time_sec must be positive")
	}

	var failureModeDistJSON, reliabilityCurveJSON, confidenceIntervalJSON []byte
	var err error

	if len(analysis.FailureModeDistribution) > 0 {
		failureModeDistJSON = analysis.FailureModeDistribution
	}
	if len(analysis.ReliabilityCurve) > 0 {
		reliabilityCurveJSON = analysis.ReliabilityCurve
	}
	if len(analysis.ConfidenceInterval) > 0 {
		confidenceIntervalJSON = analysis.ConfidenceInterval
	}

	if analysis.CreatedAt.IsZero() {
		analysis.CreatedAt = time.Now()
	}

	var id string
	query := `
		INSERT INTO magazine_reliability_analyses (
			id, crossbow_variant_code, sim_shots, sim_time_sec,
			jam_probability_per_shot, mtbf_shots, mtbf_hours,
			failure_mode_distribution, reliability_curve, confidence_interval,
			created_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		RETURNING id
	`

	err = r.db.Pool.QueryRow(context.Background(), query,
		analysis.CrossbowVariantCode,
		analysis.SimShots,
		analysis.SimTimeSec,
		analysis.JamProbabilityPerShot,
		analysis.MTBFShots,
		analysis.MTBFHours,
		failureModeDistJSON,
		reliabilityCurveJSON,
		confidenceIntervalJSON,
		analysis.CreatedAt,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to save reliability analysis: %w", err)
	}

	return id, nil
}

func (r *Repository) SaveVirtualShootSession(session *model.VirtualShootSessionRecord) error {
	if session == nil {
		return errors.New("session is nil")
	}
	if session.SessionID == "" {
		return errors.New("session_id is required")
	}
	if session.CrossbowVariantCode == "" {
		return errors.New("crossbow variant code is required")
	}

	if session.StartedAt.IsZero() {
		session.StartedAt = time.Now()
	}
	if session.UserID == "" {
		session.UserID = "anonymous"
	}

	var shotTimestamps interface{}
	if len(session.ShotTimestamps) > 0 {
		shotTimestamps = session.ShotTimestamps
	} else {
		shotTimestamps = []time.Time{}
	}

	query := `
		INSERT INTO virtual_shoot_sessions (
			session_id, crossbow_variant_code, user_id, shots_fired, jam_count,
			reload_count, elapsed_sec, average_rpm, max_instant_rpm,
			final_string_fatigue, started_at, ended_at, shot_timestamps
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
		ON CONFLICT (session_id) DO UPDATE SET
			crossbow_variant_code = EXCLUDED.crossbow_variant_code,
			user_id = EXCLUDED.user_id,
			shots_fired = EXCLUDED.shots_fired,
			jam_count = EXCLUDED.jam_count,
			reload_count = EXCLUDED.reload_count,
			elapsed_sec = EXCLUDED.elapsed_sec,
			average_rpm = EXCLUDED.average_rpm,
			max_instant_rpm = EXCLUDED.max_instant_rpm,
			final_string_fatigue = EXCLUDED.final_string_fatigue,
			started_at = EXCLUDED.started_at,
			ended_at = EXCLUDED.ended_at,
			shot_timestamps = EXCLUDED.shot_timestamps
	`

	_, err := r.db.Pool.Exec(context.Background(), query,
		session.SessionID,
		session.CrossbowVariantCode,
		session.UserID,
		session.ShotsFired,
		session.JamCount,
		session.ReloadCount,
		session.ElapsedSec,
		session.AverageRPM,
		session.MaxInstantRPM,
		session.FinalStringFatigue,
		session.StartedAt,
		session.EndedAt,
		shotTimestamps,
	)
	if err != nil {
		return fmt.Errorf("failed to save virtual shoot session: %w", err)
	}

	return nil
}

func (r *Repository) ListModernFirearms() ([]model.FirearmRecord, error) {
	query := `
		SELECT code, name, origin, intro_year, type,
			   cyclic_rate_rpm, effective_rpm, magazine_size,
			   caliber_mm, effective_range_m, muzzle_velocity_mps,
			   notes, created_at
		FROM modern_firearms
		ORDER BY intro_year ASC
	`

	rows, err := r.db.Pool.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to list modern firearms: %w", err)
	}
	defer rows.Close()

	var firearms []model.FirearmRecord
	for rows.Next() {
		var f model.FirearmRecord
		var origin, fireType, notes *string
		var introYear, cyclicRateRPM, effectiveRPM *int
		var magazineSize, effectiveRangeM *int
		var caliberMM, muzzleVelocityMPS *float64

		err := rows.Scan(
			&f.Code,
			&f.Name,
			&origin,
			&introYear,
			&fireType,
			&cyclicRateRPM,
			&effectiveRPM,
			&magazineSize,
			&caliberMM,
			&effectiveRangeM,
			&muzzleVelocityMPS,
			&notes,
			&f.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan firearm row: %w", err)
		}

		if origin != nil {
			f.Origin = *origin
		}
		if introYear != nil {
			f.IntroYear = *introYear
		}
		if fireType != nil {
			f.Type = *fireType
		}
		if cyclicRateRPM != nil {
			f.CyclicRateRPM = *cyclicRateRPM
		}
		if effectiveRPM != nil {
			f.EffectiveRPM = *effectiveRPM
		}
		if magazineSize != nil {
			f.MagazineSize = *magazineSize
		}
		if caliberMM != nil {
			f.CaliberMM = *caliberMM
		}
		if effectiveRangeM != nil {
			f.EffectiveRangeM = *effectiveRangeM
		}
		if muzzleVelocityMPS != nil {
			f.MuzzleVelocityMPS = *muzzleVelocityMPS
		}
		if notes != nil {
			f.Notes = *notes
		}

		firearms = append(firearms, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return firearms, nil
}
