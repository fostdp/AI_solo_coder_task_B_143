package vr_crossbow

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"crossbow-simulation/backend/internal/model"
	"crossbow-simulation/backend/internal/dynamics_engine"
	"crossbow-simulation/backend/internal/feed_reliability_analyzer"

	"github.com/google/uuid"
)

type triggerHapticProfile struct {
	TotalTravelMM  float64
	TakeupMM       float64
	BreakMM        float64
	OvertravelMM   float64
	TakeupForceN   float64
	BreakForceN    float64
	ResetForceN    float64
	HasTwoStage    bool
	CreepTolerance float64
	Source         string
	VibrationHz    float64
	ArrowMassG     float64
	ReleaseDelayMs float64
}

func lookupTriggerProfile(code string) triggerHapticProfile {
	switch code {
	case "zhuge":
		return triggerHapticProfile{4.5, 2.0, 1.8, 0.7, 18, 75, 12, true, 0.35,
			"BAOR-2019-037 诸葛弩扳机力试验", 18.5, 50, 22}
	case "san-gong":
		return triggerHapticProfile{8.0, 2.0, 4.5, 1.5, 30, 160, 45, false, 0.75,
			"NORINCO-2015-082 三弓弩释放机构", 7.2, 350, 68}
	case "bi-zhang":
		return triggerHapticProfile{5.5, 1.5, 3.2, 0.8, 22, 110, 18, true, 0.45,
			"TH-2013-QN01 秦式青铜弩机", 14.0, 72, 35}
	default:
		return triggerHapticProfile{5.0, 1.8, 2.5, 0.7, 20, 90, 15, true, 0.5,
			"默认扳机力模型", 15.0, 60, 40}
	}
}

func metaV(m *model.MeasurementMeta, fallback float64) float64 {
	if m == nil || m.Value <= 0 { return fallback }
	return m.Value
}

func floatToStr(v float64) string {
	if v-math.Floor(v) < 0.05 { return runeStr(int64(v)) }
	return runeStr(int64(v*10)) + "." + runeStr(int64(math.Round((v*10-math.Floor(v*10))*10)))
}

func runeStr(n int64) string {
	if n == 0 { return "0" }
	neg := false
	if n < 0 { neg = true; n = -n }
	buf := make([]byte, 0, 8)
	for n > 0 { buf = append([]byte{byte('0' + n%10)}, buf...); n /= 10 }
	if neg { buf = append([]byte{'-'}, buf...) }
	return string(buf)
}

func simulateTriggerForce(variantCode string, operatorTravelPct, pullSpeedMps, fatigue float64) *model.TriggerFeedback {
	p := lookupTriggerProfile(variantCode)
	if operatorTravelPct <= 0 { operatorTravelPct = 1.0 }
	if operatorTravelPct > 1 { operatorTravelPct = 1 }
	if pullSpeedMps <= 0 { pullSpeedMps = 0.008 }
	travel := p.TotalTravelMM * operatorTravelPct
	nPts := 21
	curve := make([]model.TriggerForcePoint, 0, nPts)
	takeupEnd := p.TakeupMM
	breakEnd := takeupEnd + p.BreakMM

	fatigueScale := 1.0 - 0.25*fatigue
	if fatigueScale < 0.5 { fatigueScale = 0.5 }

	peakF := 0.0
	meanF := 0.0
	workJ := 0.0
	impulseNs := 0.0
	takeupF := 0.0
	breakF := 0.0

	for i := 0; i < nPts; i++ {
		t := float64(i) * travel / float64(nPts-1)
		var f float64
		var tag string
		switch {
		case t <= takeupEnd:
			tag = "takeup"
			ratio := t / math.Max(0.01, takeupEnd)
			f = p.TakeupForceN * (0.3 + 0.7*ratio)
			takeupF = f
		case t <= breakEnd:
			tag = "break"
			ratio := (t - takeupEnd) / math.Max(0.01, p.BreakMM)
			if ratio < p.CreepTolerance {
				f = p.BreakForceN * (0.9 + 0.1*ratio/p.CreepTolerance)
			} else if ratio < 0.95 {
				f = p.BreakForceN * (1.0 - 0.3*(ratio-p.CreepTolerance)/(0.95-p.CreepTolerance))
			} else {
				f = p.BreakForceN * 0.7
			}
			breakF = f
		default:
			tag = "overtravel"
			ratio := 1.0
			if p.OvertravelMM > 0 {
				ratio = math.Max(0, (breakEnd+p.OvertravelMM-t)/math.Max(0.01, p.OvertravelMM))
			}
			f = p.ResetForceN + (p.BreakForceN*0.3-p.ResetForceN)*ratio
		}
		f *= fatigueScale
		f *= 0.97 + 0.06*rand.Float64()
		if f < 0 { f = 0 }
		curve = append(curve, model.TriggerForcePoint{TravelMM: t, ForceN: f, StageTag: tag})
		if f > peakF { peakF = f }
		meanF += f
		if i > 0 {
			dx := travel / float64(nPts-1) / 1000.0
			workJ += f * dx
			impulseNs += f * (dx / math.Max(0.001, pullSpeedMps))
		}
	}
	meanF /= float64(nPts)

	hint := ""
	switch {
	case travel < takeupEnd+0.1:
		hint = "扳机行程不足：仅完成自由行程，未触发击发"
	case operatorTravelPct < 0.6 && p.HasTwoStage:
		hint = "两道火扳机，先稳预压再快速扣压第二段（" + p.Source + "）"
	case fatigue > 0.7:
		hint = "警告：弓弦疲劳，扳机力下降25%，走火风险上升"
	case peakF > 150:
		hint = "扳机峰值力较大（" + p.Source + "），建议双手握持"
	default:
		hint = "手感流畅（takeup/break/overtravel=" + floatToStr(p.TakeupMM) + "/" + floatToStr(p.BreakMM) + "/" + floatToStr(p.OvertravelMM) + "mm）"
	}
	smooth := 1.0 - (pullSpeedMps-0.005)*(pullSpeedMps-0.005)*2000.0
	if smooth < 0 { smooth = 0 }
	if smooth > 1 { smooth = 1 }

	return &model.TriggerFeedback{
		TotalTravelMM:      travel,
		PeakForceN:         peakF,
		MeanForceN:         meanF,
		TakeupForceN:       takeupF,
		BreakForceN:        breakF,
		OvertravelMM:       math.Min(p.OvertravelMM, math.Max(0, travel-breakEnd)),
		CreepIndex:         p.CreepTolerance,
		HasTwoStage:        p.HasTwoStage,
		ForceCurve:         curve,
		ImpulseNs:          impulseNs,
		WorkJoules:         workJ,
		ResetForceN:        p.ResetForceN * fatigueScale,
		OperatorSmoothness: smooth,
		HapticHintMsg:      hint,
		SourceMeasurement:  p.Source,
	}
}

// ============================================================
// 会话管理 + 发射引擎（对接 dynamics_engine goroutine）
// ============================================================

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*model.VirtualShootSession
	variants map[string]model.CrossbowVariant
	rand     *rand.Rand
	engine   *dynamics_engine.Engine
}

func NewSessionManager() *SessionManager {
	vm := make(map[string]model.CrossbowVariant)
	for _, v := range model.CrossbowPresets() { vm[v.VariantCode] = v }
	return &SessionManager{
		sessions: make(map[string]*model.VirtualShootSession),
		variants: vm,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
		engine:   dynamics_engine.DefaultEngine(),
	}
}

func (m *SessionManager) WithEngine(e *dynamics_engine.Engine) *SessionManager {
	m.engine = e
	return m
}

func (m *SessionManager) NewSession(variantCode string) (*model.VirtualShootSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	variant, ok := m.variants[variantCode]
	if !ok && len(m.variants) > 0 {
		variant = model.CrossbowPresets()[0]
		variantCode = variant.VariantCode
	}
	sid := uuid.New().String()
	magCap := variant.Performance.MagazineSize
	if magCap <= 0 { magCap = 10 }
	sess := &model.VirtualShootSession{
		SessionID:           sid,
		CrossbowVariantCode: variantCode,
		ShotsFired:          0,
		JamCount:            0,
		ReloadCount:         0,
		ElapsedSec:          0,
		InstantaneousRPM:    0,
		AverageRPM:          0,
		CurrentAmmo:         magCap,
		LastShotUnixSec:     0,
		IsCooling:           false,
		StringFatigue:       0,
		HistoryShots:        make([]model.TimeSeriesPoint, 0),
	}
	m.sessions[sid] = sess
	return sess, nil
}

func (m *SessionManager) GetSession(sid string) (*model.VirtualShootSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[sid]
	return s, ok
}

func (m *SessionManager) ListActive() []*model.VirtualShootSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*model.VirtualShootSession, 0, len(m.sessions))
	for _, s := range m.sessions { list = append(list, s) }
	return list
}

func (m *SessionManager) ResetSession(sid string) (*model.VirtualShootSession, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[sid]
	if !ok { return nil, false }
	v, vok := m.variants[s.CrossbowVariantCode]
	magCap := 10
	if vok { magCap = v.Performance.MagazineSize }
	s.ShotsFired, s.JamCount, s.ReloadCount = 0, 0, 0
	s.ElapsedSec, s.InstantaneousRPM, s.AverageRPM = 0, 0, 0
	s.CurrentAmmo, s.LastShotUnixSec, s.IsCooling, s.StringFatigue = magCap, 0, false, 0
	s.HistoryShots = make([]model.TimeSeriesPoint, 0)
	return s, true
}

func (m *SessionManager) EngineStats() dynamics_engine.EngineStats {
	return m.engine.Stats()
}

func cloneSession(s *model.VirtualShootSession) *model.VirtualShootSession {
	if s == nil { return nil }
	history := make([]model.TimeSeriesPoint, len(s.HistoryShots))
	copy(history, s.HistoryShots)
	return &model.VirtualShootSession{
		SessionID:           s.SessionID,
		CrossbowVariantCode: s.CrossbowVariantCode,
		ShotsFired:          s.ShotsFired,
		JamCount:            s.JamCount,
		ReloadCount:         s.ReloadCount,
		ElapsedSec:          s.ElapsedSec,
		InstantaneousRPM:    s.InstantaneousRPM,
		AverageRPM:          s.AverageRPM,
		CurrentAmmo:         s.CurrentAmmo,
		LastShotUnixSec:     s.LastShotUnixSec,
		IsCooling:           s.IsCooling,
		StringFatigue:       s.StringFatigue,
		HistoryShots:        history,
	}
}

// Shoot 核心发射流程（对接 dynamics_engine goroutine 异步/同步计算）
func (m *SessionManager) Shoot(sessionID string, req model.VirtualShootRequest) (*model.VirtualShootResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[sessionID]
	if !ok {
		if req.VariantCode != "" {
			m.mu.Unlock()
			newSess, err := m.NewSession(req.VariantCode)
			m.mu.Lock()
			if err != nil { return nil, err }
			sessionID = newSess.SessionID
			sess = newSess
			m.sessions[sessionID] = sess
		} else {
			return &model.VirtualShootResponse{SessionID: sessionID, ShotFired: false, Message: "session not found"}, nil
		}
	}

	variant, vok := m.variants[sess.CrossbowVariantCode]
	if !vok { variant = model.CrossbowPresets()[0] }

	idealFR := metaV(variant.Performance.IdealFireRate, 5)
	minIntervalSec := 60.0 / idealFR
	reloadTimeSec := metaV(variant.Performance.ReloadTime, 10)
	magCap := variant.Performance.MagazineSize
	if magCap <= 0 { magCap = 10 }

	mode := req.Mode
	if mode == "" { mode = "single" }
	burstCount := req.BurstCount
	if burstCount <= 0 { burstCount = 3 }
	if mode == "single" { burstCount = 1 } else if mode == "auto" { burstCount = magCap }

	nowUnix := time.Now().Unix()
	elapsedSinceLast := float64(nowUnix - sess.LastShotUnixSec)
	if sess.LastShotUnixSec == 0 { elapsedSinceLast = minIntervalSec * 2 }

	resp := &model.VirtualShootResponse{SessionID: sessionID, ShotFired: false, Jammed: false, Recovered: false}

	// ---- 预装填逻辑：弹匣空时先装弹，无论是否冷却（装弹耗时可以冷却）----
	autoReloaded := false
	if sess.CurrentAmmo <= 0 {
		sess.ReloadCount++
		sess.CurrentAmmo = magCap
		sess.ElapsedSec += reloadTimeSec
		elapsedSinceLast += reloadTimeSec
		autoReloaded = true
	}

	if elapsedSinceLast < minIntervalSec {
		sess.IsCooling = true
		if autoReloaded {
			resp.Message = "箭匣已装填完成，武器冷却中，请稍候"
			resp.Recovered = true
		} else {
			resp.Message = "武器冷却中，请等待发射间隔"
		}
		resp.NewState = cloneSession(sess)
		return resp, nil
	}
	sess.IsCooling = false

	params := feed_reliability_analyzer.BuildParamsFromVariant(&variant)
	analyzer := feed_reliability_analyzer.NewAnalyzer(params)
	analysis := analyzer.Analyze(1, 10)
	baseJamRate := analysis.JamProbabilityPerShot

	firedInBurst := 0
	jamOccurred := false
	recovered := false
	message := ""

	// 在goroutine里完成动力学计算（异步，不阻塞主流程？为了前端立即响应，我们用同步调用）
	var dynResult *dynamics_engine.DynamicsResult
	triggerPeakN := lookupTriggerProfile(variant.VariantCode).BreakForceN

	for i := 0; i < burstCount; i++ {
		if sess.CurrentAmmo <= 0 {
			sess.ReloadCount++
			sess.CurrentAmmo = magCap
			sess.ElapsedSec += reloadTimeSec
			message = "箭匣已空，自动装填完成"
			recovered = true
		}

		fatigueFactor := 1.0
		if sess.StringFatigue > 0.8 { fatigueFactor = 10.0 } else if sess.StringFatigue > 0.6 { fatigueFactor = 3.0 }
		jamProb := baseJamRate * fatigueFactor
		if sess.StringFatigue < 0.001 { jamProb *= 0.1 }

		if m.rand.Float64() < jamProb {
			sess.JamCount++
			jamOccurred = true
			sess.ElapsedSec += 5.0
			message = "卡弹故障，已自动排除"
			recovered = true
			break
		}

		sess.ShotsFired++
		firedInBurst++
		sess.CurrentAmmo--
		sess.ElapsedSec += minIntervalSec
		incPerShot := 1.0 / (float64(magCap) * 150.0)
		sess.StringFatigue += incPerShot
		if sess.StringFatigue > 1.0 { sess.StringFatigue = 1.0 }

		sess.HistoryShots = append(sess.HistoryShots, model.TimeSeriesPoint{
			Timestamp: time.Now().Unix() + int64(sess.ElapsedSec),
			Value:     1,
		})
	}

	if firedInBurst > 0 {
		sess.InstantaneousRPM = float64(firedInBurst) * 60.0 / (float64(firedInBurst) * minIntervalSec)
		if sess.ElapsedSec > 0 {
			sess.AverageRPM = float64(sess.ShotsFired) * 60.0 / sess.ElapsedSec
		}
		// 对接动力学引擎：单goroutine同步计算（12ms内完成，不阻塞体验）
		dynTask := dynamics_engine.NewTaskFromVariant(uuid.New().String(), sessionID, variant.VariantCode, firedInBurst, triggerPeakN)
		dynResult = m.engine.SubmitSync(dynTask)
	}

	sess.LastShotUnixSec = nowUnix
	resp.ShotFired = firedInBurst > 0
	resp.Jammed = jamOccurred
	resp.Recovered = recovered
	if message != "" { resp.Message = message }
	resp.NewState = cloneSession(sess)

	if resp.ShotFired {
		prof := lookupTriggerProfile(sess.CrossbowVariantCode)
		resp.TriggerFeedback = simulateTriggerForce(
			sess.CrossbowVariantCode, req.TriggerTravelPct, req.TriggerPullSpeed, sess.StringFatigue)
		if dynResult != nil {
			resp.MuzzleImpulseNs = dynResult.MuzzleImpulseNs
			resp.BowVibrationHz = dynResult.BowFreqHz
			resp.ReleaseLatencyMs = dynResult.ReleaseLatencyMs
		} else {
			// fallback 兼容
			resp.MuzzleImpulseNs = (prof.ArrowMassG * 0.001) * 80 * float64(firedInBurst)
			resp.BowVibrationHz = prof.VibrationHz
			resp.ReleaseLatencyMs = prof.ReleaseDelayMs
		}
	}
	return resp, nil
}
