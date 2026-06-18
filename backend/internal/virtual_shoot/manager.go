package virtual_shoot

import (
	"math/rand"
	"sync"
	"time"

	"crossbow-simulation/backend/internal/model"
	"crossbow-simulation/backend/internal/reliability"

	"github.com/google/uuid"
)

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

	idealFireRate := variant.Performance.IdealFireRate
	if idealFireRate <= 0 {
		idealFireRate = 5
	}
	minIntervalSec := 60.0 / idealFireRate
	reloadTimeSec := variant.Performance.ReloadTime
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
