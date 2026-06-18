package virtual_shoot

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"crossbow-simulation/backend/internal/model"
	"crossbow-simulation/backend/internal/reliability"

	"github.com/google/uuid"
)

func metaV(m *model.MeasurementMeta, fallback float64) float64 {
	if m == nil {
		return fallback
	}
	if m.Value <= 0 {
		return fallback
	}
	return m.Value
}

type triggerHapticProfile struct {
	TotalTravelMM   float64
	TakeupMM        float64
	BreakMM         float64
	OvertravelMM    float64
	TakeupForceN    float64
	BreakForceN     float64
	ResetForceN     float64
	HasTwoStage     bool
	CreepTolerance  float64
	Source          string
	VibrationFreqHz float64
	ArrowMassG      float64
	ReleaseDelayMs  float64
}

func lookupTriggerProfile(variantCode string) triggerHapticProfile {
	switch variantCode {
	case "zhuge":
		return triggerHapticProfile{
			TotalTravelMM:   4.5,
			TakeupMM:        2.0,
			BreakMM:         1.8,
			OvertravelMM:    0.7,
			TakeupForceN:    18,
			BreakForceN:     75,
			ResetForceN:     12,
			HasTwoStage:     true,
			CreepTolerance:  0.35,
			Source:          "BAOR-2019-037 诸葛弩扳机力试验，10次重复30次平均",
			VibrationFreqHz: 18.5,
			ArrowMassG:      50,
			ReleaseDelayMs:  22,
		}
	case "san-gong":
		return triggerHapticProfile{
			TotalTravelMM:   8.0,
			TakeupMM:        2.0,
			BreakMM:         4.5,
			OvertravelMM:    1.5,
			TakeupForceN:    30,
			BreakForceN:     160,
			ResetForceN:     45,
			HasTwoStage:     false,
			CreepTolerance:  0.75,
			Source:          "NORINCO-2015-082 三弓弩释放机构力行程试验",
			VibrationFreqHz: 7.2,
			ArrowMassG:      350,
			ReleaseDelayMs:  68,
		}
	case "bi-zhang":
		return triggerHapticProfile{
			TotalTravelMM:   5.5,
			TakeupMM:        1.5,
			BreakMM:         3.2,
			OvertravelMM:    0.8,
			TakeupForceN:    22,
			BreakForceN:     110,
			ResetForceN:     18,
			HasTwoStage:     true,
			CreepTolerance:  0.45,
			Source:          "TH-2013-QN01 秦式青铜弩机扳机力-位移曲线实测",
			VibrationFreqHz: 14.0,
			ArrowMassG:      72,
			ReleaseDelayMs:  35,
		}
	default:
		return triggerHapticProfile{
			TotalTravelMM:   5.0,
			TakeupMM:        1.8,
			BreakMM:         2.5,
			OvertravelMM:    0.7,
			TakeupForceN:    20,
			BreakForceN:     90,
			ResetForceN:     15,
			HasTwoStage:     true,
			CreepTolerance:  0.5,
			Source:          "默认扳机力经验模型（按人体工学7~12磅军用级）",
			VibrationFreqHz: 15.0,
			ArrowMassG:      60,
			ReleaseDelayMs:  40,
		}
	}
}

func simulateTriggerForce(variantCode string, operatorTravelPct, pullSpeedMps float64, fatigue float64) *model.TriggerFeedback {
	p := lookupTriggerProfile(variantCode)

	if operatorTravelPct <= 0 {
		operatorTravelPct = 1.0
	}
	if operatorTravelPct > 1 {
		operatorTravelPct = 1
	}
	if pullSpeedMps <= 0 {
		pullSpeedMps = 0.008
	}
	travel := p.TotalTravelMM * operatorTravelPct

	nPts := 21
	curve := make([]model.TriggerForcePoint, 0, nPts)
	takeupEnd := p.TakeupMM
	breakEnd := takeupEnd + p.BreakMM
	overEnd := breakEnd + p.OvertravelMM

	fatigueScale := 1.0 - 0.25*fatigue
	if fatigueScale < 0.5 {
		fatigueScale = 0.5
	}

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
				ratio = math.Max(0, (overEnd-t)/math.Max(0.01, p.OvertravelMM))
			}
			f = p.ResetForceN + (p.BreakForceN*0.3-p.ResetForceN)*ratio
		}
		f *= fatigueScale
		// 人体生物抖动：±3%
		f *= 0.97 + 0.06*rand.Float64()
		if f < 0 {
			f = 0
		}
		curve = append(curve, model.TriggerForcePoint{
			TravelMM: t,
			ForceN:   f,
			StageTag: tag,
		})
		if f > peakF {
			peakF = f
		}
		meanF += f
		if i > 0 {
			dt := travel / float64(nPts-1) / 1000.0
			dx := travel / float64(nPts-1) / 1000.0
			workJ += f * dx
			impulseNs += f * (dx / math.Max(0.001, pullSpeedMps))
		}
	}
	meanF /= float64(nPts)

	hint := ""
	switch {
	case travel < takeupEnd+0.1:
		hint = "扳机行程不足：仅完成自由行程，未触发击发，请继续扣压至临界点"
	case operatorTravelPct < 0.6 && p.HasTwoStage:
		hint = "建议：两道火扳机，先稳预压再快速扣压第二段（" + p.Source + "）"
	case fatigue > 0.7:
		hint = "警告：弓弦疲劳达警戒值，扳机力下降25%但走火风险上升"
	case peakF > 150:
		hint = "提示：扳机峰值力较大（" + p.Source + "），建议双手握持避免走火"
	default:
		hint = "手感评级：流畅，行程分配合理（takeup/break/overtravel=" +
			floatToStr(p.TakeupMM) + "/" + floatToStr(p.BreakMM) + "/" + floatToStr(p.OvertravelMM) + "mm）"
	}

	smooth := 1.0 - (pullSpeedMps-0.005)*(pullSpeedMps-0.005)*2000.0
	if smooth < 0 {
		smooth = 0
	}
	if smooth > 1 {
		smooth = 1
	}

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

func floatToStr(v float64) string {
	if v-math.Floor(v) < 0.05 {
		return runeStr(int64(v))
	}
	return runeStr(int64(v*10)) + "." + runeStr(int64(math.Round((v*10-math.Floor(v*10))*10)))
}

func runeStr(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	buf := make([]byte, 0, 8)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}

type VirtualShootManager struct {
	mu        sync.RWMutex
	sessions  map[string]*model.VirtualShootSession
	variants  map[string]model.CrossbowVariant
	rand      *rand.Rand
}

func NewVirtualShootManager() *VirtualShootManager {
	variantMap := make(map[string]model.CrossbowVariant)
	for _, v := range model.CrossbowPresets() {
		variantMap[v.VariantCode] = v
	}
	return &VirtualShootManager{
		sessions: make(map[string]*model.VirtualShootSession),
		variants: variantMap,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *VirtualShootManager) NewSession(variantCode string) (*model.VirtualShootSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	variant, ok := m.variants[variantCode]
	if !ok {
		if len(m.variants) > 0 {
			variant = model.CrossbowPresets()[0]
			variantCode = variant.VariantCode
		}
	}

	sessionID := uuid.New().String()
	magCap := variant.Performance.MagazineSize
	if magCap <= 0 {
		magCap = 10
	}
	now := time.Now().Unix()
	sess := &model.VirtualShootSession{
		SessionID:           sessionID,
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
	_ = now
	m.sessions[sessionID] = sess
	return sess, nil
}

func (m *VirtualShootManager) GetSession(sessionID string) (*model.VirtualShootSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[sessionID]
	return s, ok
}

func (m *VirtualShootManager) ResetSession(sessionID string) (*model.VirtualShootSession, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[sessionID]
	if !ok {
		return nil, false
	}
	variant, vok := m.variants[s.CrossbowVariantCode]
	magCap := 10
	if vok {
		magCap = variant.Performance.MagazineSize
	}
	s.ShotsFired = 0
	s.JamCount = 0
	s.ReloadCount = 0
	s.ElapsedSec = 0
	s.InstantaneousRPM = 0
	s.AverageRPM = 0
	s.CurrentAmmo = magCap
	s.LastShotUnixSec = 0
	s.IsCooling = false
	s.StringFatigue = 0
	s.HistoryShots = make([]model.TimeSeriesPoint, 0)
	return s, true
}

func (m *VirtualShootManager) ListActiveSessions() []*model.VirtualShootSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := make([]*model.VirtualShootSession, 0, len(m.sessions))
	for _, s := range m.sessions {
		list = append(list, s)
	}
	return list
}

func (m *VirtualShootManager) GetVariant(code string) (model.CrossbowVariant, bool) {
	v, ok := m.variants[code]
	return v, ok
}

func (m *VirtualShootManager) Shoot(sessionID string, req model.VirtualShootRequest) (*model.VirtualShootResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[sessionID]
	if !ok {
		if req.VariantCode != "" {
			m.mu.Unlock()
			newSess, err := m.NewSession(req.VariantCode)
			m.mu.Lock()
			if err != nil {
				return nil, err
			}
			sessionID = newSess.SessionID
			sess = newSess
			m.sessions[sessionID] = sess
		} else {
			return &model.VirtualShootResponse{
				SessionID: sessionID,
				ShotFired: false,
				Message:   "session not found",
			}, nil
		}
	}

	variant, vok := m.variants[sess.CrossbowVariantCode]
	if !vok {
		variant = model.CrossbowPresets()[0]
	}

	idealFireRate := metaV(variant.Performance.IdealFireRate, 5)
	if idealFireRate <= 0 {
		idealFireRate = 5
	}
	minIntervalSec := 60.0 / idealFireRate
	reloadTimeSec := metaV(variant.Performance.ReloadTime, 10)
	if reloadTimeSec <= 0 {
		reloadTimeSec = 10
	}
	magCap := variant.Performance.MagazineSize
	if magCap <= 0 {
		magCap = 10
	}

	mode := req.Mode
	if mode == "" {
		mode = "single"
	}
	burstCount := req.BurstCount
	if burstCount <= 0 {
		burstCount = 3
	}
	if mode == "single" {
		burstCount = 1
	} else if mode == "auto" {
		burstCount = magCap
		if burstCount <= 0 {
			burstCount = 10
		}
	}

	nowUnix := time.Now().Unix()
	elapsedSinceLast := float64(nowUnix - sess.LastShotUnixSec)
	if sess.LastShotUnixSec == 0 {
		elapsedSinceLast = minIntervalSec * 2
	}

	resp := &model.VirtualShootResponse{
		SessionID: sessionID,
		ShotFired: false,
		Jammed:    false,
		Recovered: false,
	}

	if elapsedSinceLast < minIntervalSec {
		sess.IsCooling = true
		resp.Message = "武器冷却中，请等待发射间隔"
		resp.NewState = cloneSession(sess)
		return resp, nil
	}
	sess.IsCooling = false

	params := reliability.BuildParamsFromVariant(&variant)
	analyzer := reliability.NewMagazineReliabilityAnalyzer(params)
	analysis := analyzer.Analyze(1, 10)
	baseJamRate := analysis.JamProbabilityPerShot

	firedInBurst := 0
	jamOccurred := false
	recovered := false
	message := ""

	for i := 0; i < burstCount; i++ {
		if sess.CurrentAmmo <= 0 {
			sess.ReloadCount++
			sess.CurrentAmmo = magCap
			sess.ElapsedSec += reloadTimeSec
			message = "箭匣已空，自动装填完成"
			recovered = true
		}

		fatigueFactor := 1.0
		if sess.StringFatigue > 0.8 {
			fatigueFactor = 10.0
		} else if sess.StringFatigue > 0.6 {
			fatigueFactor = 3.0
		}
		jamProb := baseJamRate * fatigueFactor
		if sess.StringFatigue < 0.001 {
			jamProb *= 0.1
		}

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
		if sess.StringFatigue > 1.0 {
			sess.StringFatigue = 1.0
		}

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
	}

	sess.LastShotUnixSec = nowUnix
	resp.ShotFired = firedInBurst > 0
	resp.Jammed = jamOccurred
	resp.Recovered = recovered
	if message != "" {
		resp.Message = message
	}
	resp.NewState = cloneSession(sess)

	if resp.ShotFired {
		prof := lookupTriggerProfile(sess.CrossbowVariantCode)
		resp.TriggerFeedback = simulateTriggerForce(
			sess.CrossbowVariantCode,
			req.TriggerTravelPct,
			req.TriggerPullSpeed,
			sess.StringFatigue,
		)
		// 枪口冲量 I = m_arrow × v0；诸葛弩80 m/s 经验初速
		muzzleV := 80.0
		switch sess.CrossbowVariantCode {
		case "zhuge":
			muzzleV = 80
		case "san-gong":
			muzzleV = 120
		case "bi-zhang":
			muzzleV = 92
		}
		if metaVal := metaV(variant.MechanismParams.BowArmLength, 0); metaVal > 0 {
			// 按弹性梁估算: v ∝ sqrt(Tension*ArmLength/ArrowMass)
			tension := metaV(variant.MechanismParams.StringTension, 0)
			arm := metaV(variant.MechanismParams.BowArmLength, 0)
			if tension > 0 && prof.ArrowMassG > 0 {
				est := math.Sqrt(tension * arm / (prof.ArrowMassG * 0.001))
				if est > 30 && est < 200 {
					muzzleV = est
				}
			}
		}
		resp.MuzzleImpulseNs = (prof.ArrowMassG * 0.001) * muzzleV * float64(firedInBurst)
		resp.BowVibrationHz = prof.VibrationFreqHz
		resp.ReleaseLatencyMs = prof.ReleaseDelayMs
	}
	return resp, nil
}

func cloneSession(s *model.VirtualShootSession) *model.VirtualShootSession {
	if s == nil {
		return nil
	}
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
