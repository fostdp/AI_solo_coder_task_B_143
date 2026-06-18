package dynamics_engine

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewEngineDefaults(t *testing.T) {
	e := NewEngine(0, 0)
	if e == nil {
		t.Fatal("NewEngine returned nil")
	}
	if e.workers != 2 {
		t.Errorf("default workers should be 2, got %d", e.workers)
	}
	stats := e.Stats()
	if stats.Workers != 2 {
		t.Errorf("Stats().Workers should be 2, got %d", stats.Workers)
	}
	if stats.TasksTotal != 0 || stats.TasksDone != 0 {
		t.Error("fresh engine stats should be zero")
	}
	e.Stop()
}

func TestDefaultEngineSingleton(t *testing.T) {
	e1 := DefaultEngine()
	e2 := DefaultEngine()
	if e1 != e2 {
		t.Error("DefaultEngine should be a singleton")
	}
	if e1 == nil {
		t.Fatal("DefaultEngine returned nil")
	}
}

func TestComputeDynamicsPhysicsConsistency(t *testing.T) {
	task := &DynamicsTask{
		TaskID:         "t-physics-1",
		SessionID:      "s-test",
		VariantCode:    "zhuge",
		BowArmLengthM:  0.55,
		StringTensionN: 1200,
		ArrowMassG:     50,
		BowMassKG:      1.5,
		AirDensity:     1.225,
		Gravity:        9.81,
	}
	computeDynamics(task)
	if task.Result == nil {
		t.Fatal("Result nil after compute")
	}
	r := task.Result
	if r.MuzzleVelocityMPS < 50 || r.MuzzleVelocityMPS > 200 {
		t.Errorf("unreasonable muzzle velocity: %f m/s", r.MuzzleVelocityMPS)
	}
	expectedKE := 0.5 * 0.05 * r.MuzzleVelocityMPS * r.MuzzleVelocityMPS
	if r.KineticEnergyJ < expectedKE*0.95 || r.KineticEnergyJ > expectedKE*1.05 {
		t.Errorf("KE mismatch: got %f, expected ~%f (0.5mv^2)", r.KineticEnergyJ, expectedKE)
	}
	expectedImpulse := 0.05 * r.MuzzleVelocityMPS
	if r.MuzzleImpulseNs < expectedImpulse*0.8 || r.MuzzleImpulseNs > expectedImpulse*1.2 {
		t.Errorf("impulse mismatch: got %f, expected ~%f (mv)", r.MuzzleImpulseNs, expectedImpulse)
	}
	if r.BowFreqHz < 1 || r.BowFreqHz > 500 {
		t.Errorf("unreasonable bow vibration freq: %f Hz", r.BowFreqHz)
	}
	if r.MaxRangeM < 50 || r.MaxRangeM > 800 {
		t.Errorf("unreasonable max range: %f m", r.MaxRangeM)
	}
	if r.EfficiencyPct < 20 || r.EfficiencyPct > 50 {
		t.Errorf("efficiency should be 20-50%%, got %f", r.EfficiencyPct)
	}
	if len(r.KeyFrames) != 20 {
		t.Errorf("expected 20 keyframes, got %d", len(r.KeyFrames))
	}
	for i, kf := range r.KeyFrames {
		if kf.TimeSec < 0 {
			t.Errorf("keyframe %d: negative time", i)
		}
	}
}

func TestComputeDynamicsWithZeroDefaults(t *testing.T) {
	task := &DynamicsTask{TaskID: "zero-init"}
	computeDynamics(task)
	if task.Result == nil {
		t.Fatal("zero-input task should still produce result with defaults applied")
	}
	if task.Result.MuzzleVelocityMPS <= 0 {
		t.Error("should compute positive velocity with default params")
	}
}

func TestSubmitSyncReturnsResult(t *testing.T) {
	e := NewEngine(2, 128)
	defer e.Stop()
	task := &DynamicsTask{
		TaskID:         "submit-sync-1",
		SessionID:      "s1",
		VariantCode:    "zhuge",
		BowArmLengthM:  0.55,
		StringTensionN: 900,
		ArrowMassG:     50,
		BowMassKG:      1.5,
	}
	start := time.Now()
	res := e.SubmitSync(task)
	elapsed := time.Since(start)
	if res == nil {
		t.Fatal("SubmitSync returned nil result")
	}
	if res != task.Result {
		t.Error("SubmitSync result should be same as task.Result")
	}
	if elapsed > 2*time.Second {
		t.Errorf("SubmitSync too slow: %v (expected <2s)", elapsed)
	}
}

func TestSubmitAsyncWithDone(t *testing.T) {
	e := NewEngine(2, 128)
	defer e.Stop()
	task := &DynamicsTask{
		TaskID:         "async-1",
		SessionID:      "s2",
		VariantCode:    "zhuge",
		BowArmLengthM:  0.55,
		StringTensionN: 900,
		ArrowMassG:     50,
		BowMassKG:      1.5,
		Done:           make(chan struct{}, 1),
	}
	err := e.Submit(task)
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}
	select {
	case <-task.Done:
		if task.Result == nil {
			t.Error("task done but result nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Submit async timed out")
	}
}

func TestConcurrentSubmitSync(t *testing.T) {
	e := NewEngine(3, 512)
	defer e.Stop()
	N := 100
	var wg sync.WaitGroup
	errCount := int64(0)
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(idx int) {
			defer wg.Done()
			task := &DynamicsTask{
				TaskID:         "concurrent-" + string(rune('a'+idx%26)),
				BowArmLengthM:  0.5 + float64(idx)*0.001,
				StringTensionN: 800 + float64(idx)*10,
				ArrowMassG:     40 + float64(idx)*0.5,
				BowMassKG:      1.2 + float64(idx)*0.01,
			}
			res := e.SubmitSync(task)
			if res == nil {
				atomic.AddInt64(&errCount, 1)
			}
		}(i)
	}
	wg.Wait()
	stats := e.Stats()
	if stats.TasksTotal < int64(N) {
		t.Errorf("TasksTotal=%d should be >=%d", stats.TasksTotal, N)
	}
	if stats.TasksDone < int64(N) {
		t.Errorf("TasksDone=%d should be >=%d", stats.TasksDone, N)
	}
	if errCount > 0 {
		t.Errorf("%d of %d tasks returned nil result", errCount, N)
	}
	if stats.AvgLatencyMs < 0 {
		t.Error("avg latency should not be negative")
	}
}

func TestQueueFullFallbackToSync(t *testing.T) {
	e := NewEngine(1, 2)
	defer e.Stop()
	tasks := make([]*DynamicsTask, 0, 20)
	for i := 0; i < 20; i++ {
		task := &DynamicsTask{
			TaskID:         "qfull-" + string(rune('a'+i%26)),
			BowArmLengthM:  0.5,
			StringTensionN: 1000,
			ArrowMassG:     50,
			BowMassKG:      1.5,
		}
		tasks = append(tasks, task)
		_ = e.Submit(task)
	}
	time.Sleep(100 * time.Millisecond)
	for i, tk := range tasks {
		if tk.Result == nil {
			t.Errorf("task %d had nil result (queue-full fallback should compute sync)", i)
		}
	}
	stats := e.Stats()
	if stats.TasksRejected == 0 {
		t.Log("note: no tasks were rejected even with tiny queue (acceptable)")
	}
}

func TestStopEngineGracefully(t *testing.T) {
	e := NewEngine(2, 64)
	time.Sleep(20 * time.Millisecond)
	done := make(chan struct{})
	go func() {
		e.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Stop() deadlocked")
	}
	task := &DynamicsTask{
		BowArmLengthM:  0.5,
		StringTensionN: 900,
		ArrowMassG:     50,
		BowMassKG:      1.5,
	}
	e.Submit(task)
	if task.Result == nil {
		t.Error("submit after stop should still fallback to sync compute")
	}
}

func TestKeyFrameOrdering(t *testing.T) {
	task := &DynamicsTask{
		BowArmLengthM:  0.55,
		StringTensionN: 1000,
		ArrowMassG:     50,
		BowMassKG:      1.5,
	}
	computeDynamics(task)
	kfs := task.Result.KeyFrames
	if len(kfs) < 2 {
		t.Fatal("not enough keyframes")
	}
	prevT := -1.0
	for i, kf := range kfs {
		if kf.TimeSec <= prevT && i > 0 {
			t.Errorf("keyframe %d time not increasing: %f <= %f", i, kf.TimeSec, prevT)
		}
		prevT = kf.TimeSec
	}
	lastArrowX := kfs[len(kfs)-1].ArrowX
	if lastArrowX < kfs[0].ArrowX {
		t.Error("arrow should move forward (X increasing)")
	}
}

func TestStressHighConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping high-concurrency stress in short mode")
	}
	e := NewEngine(4, 1024)
	defer e.Stop()
	N := 500
	var wg sync.WaitGroup
	var success int64
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			task := &DynamicsTask{
				TaskID:         "stress",
				BowArmLengthM:  0.5,
				StringTensionN: 1000,
				ArrowMassG:     50,
				BowMassKG:      1.5,
			}
			res := e.SubmitSync(task)
			if res != nil && res.MuzzleVelocityMPS > 0 {
				atomic.AddInt64(&success, 1)
			}
		}()
	}
	wg.Wait()
	if success < int64(float64(N)*0.99) {
		t.Errorf("only %d/%d tasks succeeded (99%% required)", success, N)
	}
	t.Logf("Stress: %d/%d succeeded, stats=%+v", success, N, e.Stats())
}
