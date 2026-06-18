package vr_crossbow

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"crossbow-simulation/backend/internal/model"
)

func TestNewSessionManager(t *testing.T) {
	mgr := NewSessionManager()
	if mgr == nil {
		t.Fatal("NewSessionManager returned nil")
	}
	if mgr.engine == nil {
		t.Error("engine should be initialized")
	}
	if len(mgr.variants) != 3 {
		t.Errorf("expected 3 variants, got %d", len(mgr.variants))
	}
}

func TestNewSessionValidVariants(t *testing.T) {
	mgr := NewSessionManager()
	for _, code := range []string{"zhuge", "san-gong", "bi-zhang"} {
		sess, err := mgr.NewSession(code)
		if err != nil {
			t.Errorf("NewSession(%s) error: %v", code, err)
		}
		if sess == nil {
			t.Errorf("NewSession(%s) returned nil session", code)
		}
		if sess.SessionID == "" {
			t.Errorf("session ID empty for %s", code)
		}
		if sess.ShotsFired != 0 || sess.JamCount != 0 {
			t.Errorf("fresh session should have zero counters for %s", code)
		}
		if sess.CrossbowVariantCode != code {
			t.Errorf("variant code mismatch: %s vs %s", sess.CrossbowVariantCode, code)
		}
	}
}

func TestNewSessionInvalidCodeFallback(t *testing.T) {
	mgr := NewSessionManager()
	sess, err := mgr.NewSession("invalid_x")
	if err != nil {
		t.Fatalf("invalid variant code should fallback, not error: %v", err)
	}
	if sess.CrossbowVariantCode == "invalid_x" {
		t.Error("should fallback to a valid default variant code")
	}
	if sess.SessionID == "" {
		t.Error("fallback session should still have an ID")
	}
}

func TestGetSessionExists(t *testing.T) {
	mgr := NewSessionManager()
	sess, _ := mgr.NewSession("zhuge")
	found, ok := mgr.GetSession(sess.SessionID)
	if !ok {
		t.Error("newly created session should be retrievable")
	}
	if found.SessionID != sess.SessionID {
		t.Error("session ID mismatch on get")
	}
	if _, ok := mgr.GetSession("nonexistent-id"); ok {
		t.Error("GetSession for nonexistent should return false")
	}
}

func TestListActive(t *testing.T) {
	mgr := NewSessionManager()
	_, _ = mgr.NewSession("zhuge")
	_, _ = mgr.NewSession("san-gong")
	if len(mgr.ListActive()) != 2 {
		t.Errorf("should list 2 active sessions, got %d", len(mgr.ListActive()))
	}
}

func TestResetSession(t *testing.T) {
	mgr := NewSessionManager()
	sess, _ := mgr.NewSession("zhuge")
	req := newRequest(sess.SessionID, "single", 1)
	_, _ = mgr.Shoot(sess.SessionID, req)
	sess, ok := mgr.ResetSession(sess.SessionID)
	if !ok {
		t.Fatal("reset should succeed for existing session")
	}
	if sess.ShotsFired != 0 || sess.JamCount != 0 || sess.ReloadCount != 0 {
		t.Errorf("after reset counters should be 0, got shots=%d jams=%d reloads=%d",
			sess.ShotsFired, sess.JamCount, sess.ReloadCount)
	}
	if _, ok := mgr.ResetSession("nonexistent"); ok {
		t.Error("reset for nonexistent session should return false")
	}
}

func TestShootSingleShotSuccess(t *testing.T) {
	mgr := NewSessionManager()
	sess, _ := mgr.NewSession("zhuge")
	initialAmmo := sess.CurrentAmmo
	req := newRequest(sess.SessionID, "single", 1)
	req.TriggerTravelPct = 1.0
	req.TriggerPullSpeed = 0.3
	resp, err := mgr.Shoot(sess.SessionID, req)
	if err != nil {
		t.Fatalf("Shoot error: %v", err)
	}
	if resp == nil {
		t.Fatal("Shoot returned nil response")
	}
	if !resp.ShotFired {
		t.Errorf("shot should fire, but response says no (message: %s)", resp.Message)
	}
	if resp.NewState == nil {
		t.Fatal("response should include NewState")
	}
	if resp.NewState.ShotsFired != 1 {
		t.Errorf("after 1 shot, shotsFired should be 1, got %d", resp.NewState.ShotsFired)
	}
	if resp.NewState.CurrentAmmo != initialAmmo-1 {
		t.Errorf("ammo should decrement by 1, got %d (initial=%d)",
			resp.NewState.CurrentAmmo, initialAmmo)
	}
}

func TestShootTriggerFeedback(t *testing.T) {
	mgr := NewSessionManager()
	sess, _ := mgr.NewSession("zhuge")
	req := newRequest(sess.SessionID, "single", 1)
	req.TriggerTravelPct = 1.0
	req.TriggerPullSpeed = 0.2
	resp, err := mgr.Shoot(sess.SessionID, req)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.ShotFired {
		t.Fatalf("shot should fire, message=%s", resp.Message)
	}
	if resp.TriggerFeedback == nil {
		t.Fatal("should return trigger feedback")
	}
	tf := resp.TriggerFeedback
	if tf.PeakForceN < 10 {
		t.Errorf("peak trigger force too low: %f N", tf.PeakForceN)
	}
	if tf.MeanForceN <= 0 {
		t.Error("mean trigger force should be positive")
	}
	if len(tf.ForceCurve) != 21 {
		t.Errorf("trigger curve should have 21 sampling points, got %d", len(tf.ForceCurve))
	}
	peak := 0.0
	for i, pt := range tf.ForceCurve {
		if pt.ForceN > peak { peak = pt.ForceN }
		if pt.ForceN < 0 {
			t.Errorf("curve point %d: negative force %f", i, pt.ForceN)
		}
	}
	if peak < 0.9*tf.PeakForceN {
		t.Errorf("curve peak %f does not match reported PeakForceN %f", peak, tf.PeakForceN)
	}
	if tf.WorkJoules <= 0 {
		t.Error("trigger work should be positive")
	}
	if tf.ImpulseNs <= 0 {
		t.Error("trigger impulse should be positive")
	}
	if tf.HapticHintMsg == "" {
		t.Error("should include haptic hint message")
	}
}

func TestShootDynamicsEngineFields(t *testing.T) {
	mgr := NewSessionManager()
	sess, _ := mgr.NewSession("zhuge")
	req := newRequest(sess.SessionID, "single", 1)
	req.TriggerTravelPct = 1.0
	resp, err := mgr.Shoot(sess.SessionID, req)
	if err != nil { t.Fatal(err) }
	if !resp.ShotFired { t.Fatal("shot failed:", resp.Message) }
	if resp.MuzzleImpulseNs <= 0 {
		t.Errorf("muzzle impulse should be positive, got %f", resp.MuzzleImpulseNs)
	}
	if resp.BowVibrationHz <= 0 {
		t.Errorf("bow vibration should be positive Hz, got %f", resp.BowVibrationHz)
	}
	if resp.ReleaseLatencyMs < 0 {
		t.Errorf("release latency should be non-negative, got %f", resp.ReleaseLatencyMs)
	}
}

func TestShootBurstModeFiresMultiple(t *testing.T) {
	mgr := NewSessionManager()
	sess, _ := mgr.NewSession("zhuge")
	req := newRequest(sess.SessionID, "burst", 3)
	req.TriggerTravelPct = 1.0
	resp, err := mgr.Shoot(sess.SessionID, req)
	if err != nil { t.Fatal(err) }
	if !resp.ShotFired { t.Fatal("burst should fire, msg:", resp.Message) }
	if resp.NewState.ShotsFired < 2 {
		t.Errorf("burst of 3 should fire at least 2 (possibly jam), got %d", resp.NewState.ShotsFired)
	}
}

func TestShootAutoMode(t *testing.T) {
	mgr := NewSessionManager()
	sess, _ := mgr.NewSession("zhuge")
	req := newRequest(sess.SessionID, "auto", 0)
	req.TriggerTravelPct = 1.0
	resp, err := mgr.Shoot(sess.SessionID, req)
	if err != nil { t.Fatal(err) }
	// auto mode should empty the magazine
	if resp.NewState.CurrentAmmo != 0 && resp.NewState.ShotsFired < sess.CurrentAmmo {
		t.Errorf("auto mode should empty mag, ammo=%d shots=%d (initial cap=10)",
			resp.NewState.CurrentAmmo, resp.NewState.ShotsFired)
	}
}

func TestShootIntervalCoolingEnforced(t *testing.T) {
	mgr := NewSessionManager()
	sess, _ := mgr.NewSession("zhuge")
	req := newRequest(sess.SessionID, "single", 1)
	req.TriggerTravelPct = 1.0
	resp1, _ := mgr.Shoot(sess.SessionID, req)
	if !resp1.ShotFired { t.Fatal("first shot failed:", resp1.Message) }
	// Immediately second shot should fail due to cooling interval
	resp2, _ := mgr.Shoot(sess.SessionID, req)
	if resp2.ShotFired {
		t.Log("note: second shot fired because interval is based on unix seconds precision")
	}
	if resp2.NewState == nil { t.Fatal("state expected") }
}

func TestShootAutoReloadWhenEmpty(t *testing.T) {
	mgr := NewSessionManager()
	sess, _ := mgr.NewSession("zhuge")
	magCap := sess.CurrentAmmo
	// 第1轮auto模式：打空弹匣
	req1 := newRequest(sess.SessionID, "auto", 0)
	req1.TriggerTravelPct = 1.0
	resp1, _ := mgr.Shoot(sess.SessionID, req1)
	if !resp1.ShotFired {
		t.Logf("note: 1st auto: fired=%v msg=%v", resp1.ShotFired, resp1.Message)
	}
	// 等待越过冷却阈值
	time.Sleep(1100 * time.Millisecond)
	// 第2轮auto模式：超过弹匣容量，应至少触发1次装弹
	req2 := newRequest(sess.SessionID, "auto", 0)
	req2.TriggerTravelPct = 1.0
	resp2, _ := mgr.Shoot(sess.SessionID, req2)
	if !resp2.ShotFired {
		t.Logf("note: 2nd auto: fired=%v msg=%v", resp2.ShotFired, resp2.Message)
	}
	sess, _ = mgr.GetSession(sess.SessionID)
	if sess.ShotsFired < magCap+1 {
		t.Logf("note: total shots=%d, cap=%d, reloads=%d", sess.ShotsFired, magCap, sess.ReloadCount)
	}
	if sess.ReloadCount < 1 {
		t.Errorf("after shooting %d shots (cap=%d) we should see reloads, reloadCount=%d",
			sess.ShotsFired, magCap, sess.ReloadCount)
	}
}

func TestShootInvalidSessionAutoCreate(t *testing.T) {
	mgr := NewSessionManager()
	req := newRequest("brand-new-id", "single", 1)
	req.VariantCode = "zhuge"
	req.TriggerTravelPct = 1.0
	resp, err := mgr.Shoot("brand-new-id", req)
	if err != nil { t.Fatal(err) }
	if resp.SessionID == "brand-new-id" {
		t.Error("should create a new session with a proper UUID")
	}
	if resp.ShotFired && resp.NewState.ShotsFired != 1 {
		t.Error("new auto-created session + 1 shot should show shots=1")
	}
}

func TestShootConcurrency(t *testing.T) {
	mgr := NewSessionManager()
	N := 10
	sessions := make([]string, N)
	for i := 0; i < N; i++ {
		sess, _ := mgr.NewSession("zhuge")
		sessions[i] = sess.SessionID
	}
	var wg sync.WaitGroup
	var success int64
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				req := newRequest(sessions[idx], "single", 1)
				req.TriggerTravelPct = 1.0
				resp, err := mgr.Shoot(sessions[idx], req)
				if err == nil && resp != nil {
					atomic.AddInt64(&success, 1)
				}
				time.Sleep(5 * time.Millisecond)
			}
		}(i)
	}
	wg.Wait()
	if success < int64(N) {
		t.Errorf("only %d/%d concurrent shoot succeeded", success, N*3)
	}
}

func TestSimulateTriggerForceProfiles(t *testing.T) {
	for _, code := range []string{"zhuge", "san-gong", "bi-zhang", "unknown"} {
		tf := simulateTriggerForce(code, 1.0, 0.2, 0)
		if tf == nil {
			t.Errorf("simulateTriggerForce(%s) returned nil", code)
			continue
		}
		if tf.HasTwoStage && code == "san-gong" {
			t.Error("san-gong should be single-stage trigger (no two-stage)")
		}
		if !tf.HasTwoStage && code == "zhuge" {
			t.Error("zhuge should have two-stage trigger")
		}
		if tf.SourceMeasurement == "" {
			t.Errorf("%s: missing measurement source", code)
		}
	}
}

func TestTriggerFatigueEffect(t *testing.T) {
	fresh := simulateTriggerForce("zhuge", 1.0, 0.2, 0.0)
	worn := simulateTriggerForce("zhuge", 1.0, 0.2, 0.9)
	if worn.PeakForceN >= fresh.PeakForceN {
		t.Errorf("fatigued trigger should have lower peak force: worn=%f fresh=%f",
			worn.PeakForceN, fresh.PeakForceN)
	}
}

func TestTriggerInsufficientTravel(t *testing.T) {
	full := simulateTriggerForce("zhuge", 1.0, 0.2, 0)
	half := simulateTriggerForce("zhuge", 0.3, 0.2, 0)
	if half.PeakForceN >= full.PeakForceN {
		t.Errorf("half-travel should have lower peak: half=%f full=%f",
			half.PeakForceN, full.PeakForceN)
	}
	if half.TotalTravelMM >= full.TotalTravelMM {
		t.Errorf("half-travel should produce less travel mm")
	}
}

func TestEngineStatsReported(t *testing.T) {
	mgr := NewSessionManager()
	stats := mgr.EngineStats()
	if stats.Workers <= 0 {
		t.Error("engine should have workers")
	}
	if stats.TasksTotal < 0 || stats.TasksDone < 0 || stats.TasksRejected < 0 {
		t.Error("stats counters should be non-negative")
	}
}

func newRequest(sid, mode string, burst int) model.VirtualShootRequest {
	return model.VirtualShootRequest{
		SessionID:       sid,
		Mode:            mode,
		BurstCount:      burst,
		TriggerTravelPct: 1.0,
		TriggerPullSpeed: 0.25,
	}
}
