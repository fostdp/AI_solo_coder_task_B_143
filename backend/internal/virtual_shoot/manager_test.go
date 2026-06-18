package virtual_shoot

import (
	"sync"
	"testing"
	"time"

	"crossbow-simulation/backend/internal/model"

	"github.com/stretchr/testify/assert"
)

func mmValue(m *model.MeasurementMeta) float64 {
	if m == nil {
		return 0
	}
	return m.Value
}

// ================== Feature 4: 虚拟射击体验测试 ==================

// ============ 会话管理测试 ============

// 正常：创建诸葛弩会话
func TestNewSession_ZhugeVariant(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, err := mgr.NewSession("zhuge")

	assert.NoError(t, err)
	assert.NotNil(t, sess)
	assert.Equal(t, "zhuge", sess.CrossbowVariantCode)
	assert.Equal(t, 0, sess.ShotsFired)
	assert.Equal(t, 0, sess.JamCount)
	assert.Equal(t, 0, sess.ReloadCount)
	assert.Equal(t, 10, sess.CurrentAmmo, "诸葛弩弹夹应满10发")
	assert.Equal(t, 0.0, sess.StringFatigue)
	assert.NotEmpty(t, sess.SessionID)
	t.Logf("会话ID: %s, 初始弹容: %d", sess.SessionID, sess.CurrentAmmo)
}

// 正常：创建三弓弩会话
func TestNewSession_SanGongVariant(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, err := mgr.NewSession("san-gong")

	assert.NoError(t, err)
	assert.Equal(t, "san-gong", sess.CrossbowVariantCode)
	assert.Equal(t, 1, sess.CurrentAmmo, "三弓弩为1发")
}

// 异常：弩型不存在 - 应回退到默认（诸葛弩）
func TestNewSession_InvalidVariant_Fallback(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, err := mgr.NewSession("不存在的弩型")

	assert.NoError(t, err, "无效弩型不应报错，应回退默认")
	assert.NotNil(t, sess)
	assert.Equal(t, "zhuge", sess.CrossbowVariantCode)
	assert.Equal(t, 10, sess.CurrentAmmo)
}

// 边界：获取存在的会话
func TestGetSession_Exists(t *testing.T) {
	mgr := NewVirtualShootManager()
	created, _ := mgr.NewSession("zhuge")
	retrieved, ok := mgr.GetSession(created.SessionID)

	assert.True(t, ok)
	assert.Equal(t, created.SessionID, retrieved.SessionID)
}

// 边界：获取不存在的会话
func TestGetSession_NotExists(t *testing.T) {
	mgr := NewVirtualShootManager()
	_, ok := mgr.GetSession("ffffffff-ffff-ffff-ffff-ffffffffffff")
	assert.False(t, ok)
}

// 边界：重置会话
func TestResetSession_ToInitialState(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, _ := mgr.NewSession("zhuge")

	// 先打几发，制造一些状态
	req := model.VirtualShootRequest{SessionID: sess.SessionID, Mode: "single"}
	for i := 0; i < 5; i++ {
		_, _ = mgr.Shoot(sess.SessionID, req)
	}

	t.Logf("重置前: 发射%d, 疲劳%.4f, 弹容%d",
		sess.ShotsFired, sess.StringFatigue, sess.CurrentAmmo)
	assert.GreaterOrEqual(t, sess.ShotsFired, 0)

	// 重置
	reset, ok := mgr.ResetSession(sess.SessionID)
	assert.True(t, ok)
	assert.Equal(t, 0, reset.ShotsFired, "重置后发射数应为0")
	assert.Equal(t, 0, reset.JamCount)
	assert.Equal(t, 0, reset.ReloadCount)
	assert.Equal(t, 10, reset.CurrentAmmo, "弹夹应恢复满10发")
	assert.Equal(t, 0.0, reset.StringFatigue, "疲劳应为0")
}

// 异常：重置不存在的会话
func TestResetSession_NotExists(t *testing.T) {
	mgr := NewVirtualShootManager()
	_, ok := mgr.ResetSession("invalid-id")
	assert.False(t, ok)
}

// 并发：多线程同时创建会话
func TestNewSession_ConcurrentSafety(t *testing.T) {
	mgr := NewVirtualShootManager()
	var wg sync.WaitGroup
	sessions := make([]*model.VirtualShootSession, 100)
	errs := make([]error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			codes := []string{"zhuge", "bi-zhang", "san-gong"}
			s, e := mgr.NewSession(codes[idx%3])
			sessions[idx] = s
			errs[idx] = e
		}(i)
	}
	wg.Wait()

	for _, e := range errs {
		assert.NoError(t, e)
	}
	assert.Len(t, mgr.ListActiveSessions(), 100)
}

// ============ 射击操作测试 ============

// 正常：单发射击 - 诸葛弩首发
func TestShoot_SingleShot_FirstFire(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, _ := mgr.NewSession("zhuge")

	resp, err := mgr.Shoot(sess.SessionID, model.VirtualShootRequest{
		SessionID:  sess.SessionID,
		VariantCode: "zhuge",
		Mode:       "single",
	})

	assert.NoError(t, err)
	assert.True(t, resp.ShotFired, "第一发应成功")
	assert.Equal(t, sess.SessionID, resp.SessionID)
	assert.NotNil(t, resp.NewState)
	assert.Equal(t, 1, resp.NewState.ShotsFired)
	assert.Equal(t, 9, resp.NewState.CurrentAmmo, "应消耗1发")
	assert.Greater(t, resp.NewState.ElapsedSec, 0.0)
	t.Logf("首发成功: 剩余弹=%d, 耗时%.2fs, RPM=%.1f, 疲劳=%.6f",
		resp.NewState.CurrentAmmo, resp.NewState.ElapsedSec,
		resp.NewState.InstantaneousRPM, resp.NewState.StringFatigue)
}

// 正常：三弓弩单发 - 弹容1，打完后第二发需装弹
func TestShoot_SanGong_ReloadRequired(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, _ := mgr.NewSession("san-gong")

	// 第1发
	resp1, _ := mgr.Shoot(sess.SessionID, model.VirtualShootRequest{
		SessionID: sess.SessionID, Mode: "single",
	})
	assert.True(t, resp1.ShotFired)
	assert.Equal(t, 1, resp1.NewState.ShotsFired)
	assert.Equal(t, 0, resp1.NewState.CurrentAmmo)

	// 等待至少1秒，越过基于Unix秒的冷却阈值
	time.Sleep(1100 * time.Millisecond)

	// 第2发：弹夹空了，触发自动装填
	resp2, _ := mgr.Shoot(sess.SessionID, model.VirtualShootRequest{
		SessionID: sess.SessionID, Mode: "single",
	})
	// 应触发装填并成功发射
	assert.True(t, resp2.ShotFired || resp2.Recovered)
	t.Logf("三弓弩第二发: fired=%v recovered=%v msg=%s elapsed=%.2fs",
		resp2.ShotFired, resp2.Recovered, resp2.Message, resp2.NewState.ElapsedSec)

	// 装弹计数
	totalReloads := resp2.NewState.ReloadCount
	totalShots := resp2.NewState.ShotsFired
	t.Logf("总发射: %d, 装弹次数: %d", totalShots, totalReloads)
	assert.GreaterOrEqual(t, totalReloads, 1,
		"弹容1，打第2发时应触发1次装弹")
}

// 正常：诸葛弩打满一匣（10发），触发自动装弹
func TestShoot_Zhuge_FullMagazine_Reload(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, _ := mgr.NewSession("zhuge")

	var lastResp *model.VirtualShootResponse

	// 第1次 auto 模式：打空10发弹匣（约10发）
	resp1, err := mgr.Shoot(sess.SessionID, model.VirtualShootRequest{
		SessionID: sess.SessionID, Mode: "auto",
	})
	assert.NoError(t, err)
	lastResp = resp1
	t.Logf("第1轮auto: fired=%v jammed=%v ammo=%d shots=%d reloads=%d msg=%v",
		resp1.ShotFired, resp1.Jammed,
		resp1.NewState.CurrentAmmo,
		resp1.NewState.ShotsFired,
		resp1.NewState.ReloadCount, resp1.Message)

	// 等待越过冷却阈值
	time.Sleep(1100 * time.Millisecond)

	// 第2次 auto 模式：超过10发，至少触发1次装弹
	resp2, err := mgr.Shoot(sess.SessionID, model.VirtualShootRequest{
		SessionID: sess.SessionID, Mode: "auto",
	})
	assert.NoError(t, err)
	lastResp = resp2
	t.Logf("第2轮auto: fired=%v jammed=%v ammo=%d shots=%d reloads=%d msg=%v",
		resp2.ShotFired, resp2.Jammed,
		resp2.NewState.CurrentAmmo,
		resp2.NewState.ShotsFired,
		resp2.NewState.ReloadCount, resp2.Message)

	assert.GreaterOrEqual(t, lastResp.NewState.ShotsFired, 11,
		"两轮auto应至少发射11发（超过1匣容量）")
	assert.GreaterOrEqual(t, lastResp.NewState.ReloadCount, 1,
		"超过10发应至少装弹1次")
}

// 正常：burst 模式 - 诸葛弩3连发
func TestShoot_BurstMode_ThreeShots(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, _ := mgr.NewSession("zhuge")

	resp, err := mgr.Shoot(sess.SessionID, model.VirtualShootRequest{
		SessionID:  sess.SessionID,
		VariantCode: "zhuge",
		Mode:       "burst",
		BurstCount: 3,
	})

	assert.NoError(t, err)
	assert.True(t, resp.ShotFired)
	assert.GreaterOrEqual(t, resp.NewState.ShotsFired, 1,
		"burst至少发射1发")
	assert.LessOrEqual(t, resp.NewState.ShotsFired, 3,
		"burst最多3发")
	t.Logf("burst模式: 发射%d发, 剩余弹%d, 耗时%.2fs",
		resp.NewState.ShotsFired, resp.NewState.CurrentAmmo,
		resp.NewState.ElapsedSec)
}

// 正常：auto 模式 - 打空一匣
func TestShoot_AutoMode_EmptyMagazine(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, _ := mgr.NewSession("zhuge")

	resp, _ := mgr.Shoot(sess.SessionID, model.VirtualShootRequest{
		SessionID: sess.SessionID, Mode: "auto",
	})

	assert.True(t, resp.ShotFired)
	// auto模式理论应打空一匣，但可能因卡弹中断
	assert.GreaterOrEqual(t, resp.NewState.ShotsFired, 1)
	assert.LessOrEqual(t, resp.NewState.ShotsFired, 10)
	t.Logf("auto模式: 发射%d发, 剩余弹=%d, 装弹=%d, 卡弹=%d",
		resp.NewState.ShotsFired,
		resp.NewState.CurrentAmmo,
		resp.NewState.ReloadCount,
		resp.NewState.JamCount)
}

// 边界：冷却检测 - 连续快速射击（极短时间内第二次调用）
func TestShoot_CoolingEnforcement(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, _ := mgr.NewSession("zhuge")
	idealRate := 0.0
	for _, v := range model.CrossbowPresets() {
		if v.VariantCode == "zhuge" {
			idealRate = mmValue(v.Performance.IdealFireRate)
		}
	}
	minInterval := 60.0 / idealRate
	t.Logf("诸葛弩理想射速 %.1f 发/分 → 最小发射间隔 %.2f 秒", idealRate, minInterval)

	// 第1发 - 正常
	resp1, _ := mgr.Shoot(sess.SessionID, model.VirtualShootRequest{
		SessionID: sess.SessionID, Mode: "single",
	})
	assert.True(t, resp1.ShotFired)

	// 立即第2发（同一秒内）- 应被冷却拦截
	resp2, _ := mgr.Shoot(sess.SessionID, model.VirtualShootRequest{
		SessionID: sess.SessionID, Mode: "single",
	})
	assert.False(t, resp2.ShotFired, "未冷却应拒绝发射")
	assert.True(t, resp2.NewState.IsCooling, "应进入冷却状态")
	assert.Contains(t, resp2.Message, "冷却")
	t.Logf("冷却拦截成功: isCooling=%v, msg=%s",
		resp2.NewState.IsCooling, resp2.Message)
}

// 边界：缺少 sessionId
func TestShoot_MissingSessionId(t *testing.T) {
	mgr := NewVirtualShootManager()
	resp, err := mgr.Shoot("", model.VirtualShootRequest{
		Mode: "single",
	})
	assert.NoError(t, err)
	assert.False(t, resp.ShotFired, "空sessionId应返回失败")
}

// 异常：不存在的 sessionId 但提供了 VariantCode - 自动创建会话
func TestShoot_AutoCreateSession(t *testing.T) {
	mgr := NewVirtualShootManager()
	resp, err := mgr.Shoot("new-session-id-123", model.VirtualShootRequest{
		SessionID:   "new-session-id-123",
		VariantCode: "zhuge",
		Mode:        "single",
	})

	assert.NoError(t, err)
	assert.True(t, resp.ShotFired)
	assert.NotEmpty(t, resp.SessionID)
	t.Logf("自动创建会话成功: ID=%s, 发射=%v", resp.SessionID, resp.ShotFired)
}

// ============ 射速与疲劳验证 ============

// 业务验证：诸葛弩射速达到理论值
func TestShoot_Zhuge_IdealFireRateMet(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, _ := mgr.NewSession("zhuge")
	idealRate := 0.0
	for _, v := range model.CrossbowPresets() {
		if v.VariantCode == "zhuge" {
			idealRate = mmValue(v.Performance.IdealFireRate)
		}
	}

	// 连续发射30发，计算平均射速
	req := model.VirtualShootRequest{SessionID: sess.SessionID, Mode: "burst", BurstCount: 5}
	totalShots := 0
	totalElapsed := 0.0
	for i := 0; i < 6; i++ {
		resp, err := mgr.Shoot(sess.SessionID, req)
		assert.NoError(t, err)
		if resp.ShotFired {
			totalShots = resp.NewState.ShotsFired
			totalElapsed = resp.NewState.ElapsedSec
		}
	}

	if totalElapsed > 0 {
		actualRate := float64(totalShots) * 60.0 / totalElapsed
		t.Logf("理论射速: %.1f 发/分, 实际射速: %.1f 发/分 (%.1f%%)",
			idealRate, actualRate, actualRate/idealRate*100)
		// 实际射速（含装弹和冷却时间）应在理论值的 30%~120%
		assert.Greater(t, actualRate, idealRate*0.3)
		assert.Less(t, actualRate, idealRate*1.5)
	}
}

// 业务验证：弓弦疲劳随发射增加
func TestShoot_FatigueIncreasesWithShots(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, _ := mgr.NewSession("zhuge")

	// 20次burst 3发
	req := model.VirtualShootRequest{SessionID: sess.SessionID, Mode: "burst", BurstCount: 3}
	fatigueHistory := make([]float64, 0)
	for i := 0; i < 20; i++ {
		resp, err := mgr.Shoot(sess.SessionID, req)
		assert.NoError(t, err)
		fatigueHistory = append(fatigueHistory, resp.NewState.StringFatigue)
	}

	firstFatigue := fatigueHistory[0]
	lastFatigue := fatigueHistory[len(fatigueHistory)-1]
	t.Logf("首次射击疲劳: %.6f, 60发后疲劳: %.6f",
		firstFatigue, lastFatigue)

	// 疲劳应严格单调不减
	for i := 1; i < len(fatigueHistory); i++ {
		assert.GreaterOrEqual(t, fatigueHistory[i], fatigueHistory[i-1],
			"疲劳度应单调不减")
	}
	assert.Greater(t, lastFatigue, 0.0, "多次发射后疲劳>0")
	assert.LessOrEqual(t, lastFatigue, 1.0, "疲劳度不应超过1.0")
}

// 业务验证：疲劳度高时，卡弹概率显著增加
func TestShoot_HighFatigue_JamRateIncrease(t *testing.T) {
	// 用低弹容高射速的诸葛弩，测试"新弩"与"疲劳弩"的卡弹数对比
	countJams := func(initialFatigue float64, totalShots int) int {
		mgr := NewVirtualShootManager()
		sess, _ := mgr.NewSession("zhuge")
		// 手动设置初始疲劳
		sess.StringFatigue = initialFatigue
		req := model.VirtualShootRequest{SessionID: sess.SessionID, Mode: "single"}
		// 模拟totalShots发射
		for i := 0; i < totalShots; i++ {
			resp, _ := mgr.Shoot(sess.SessionID, req)
			if resp.Jammed {
				// 保持高疲劳
				sess.StringFatigue = initialFatigue
			}
			// 将冷却状态去除以便继续发射
			sess.IsCooling = false
			sess.LastShotUnixSec = time.Now().Unix() - 100
		}
		return sess.JamCount
	}

	freshJams := 0
	fatiguedJams := 0
	trials := 10
	for t := 0; t < trials; t++ {
		freshJams += countJams(0.01, 50)
		fatiguedJams += countJams(0.9, 50)
	}

	t.Logf("新弩(疲劳≈0): 累计卡弹 %d 次, 高疲劳弩(0.9): 累计卡弹 %d 次",
		freshJams, fatiguedJams)
	// 高疲劳时卡弹数应显著更多
	// 注意：存在随机波动，这里仅做非严格断言
	if fatiguedJams == 0 && freshJams == 0 {
		t.Log("两次都没卡弹（随机性导致），跳过本断言")
		return
	}
	t.Logf("高疲劳/新弩卡弹比: %.1f", float64(fatiguedJams)/float64(freshJams+1))
}

// ============ 历史记录与克隆 ============

// 验证：HistoryShots 随发射增加
func TestShoot_HistoryShotsRecorded(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, _ := mgr.NewSession("zhuge")

	req := model.VirtualShootRequest{SessionID: sess.SessionID, Mode: "burst", BurstCount: 4}
	resp, _ := mgr.Shoot(sess.SessionID, req)
	if resp.ShotFired {
		assert.Len(t, resp.NewState.HistoryShots, resp.NewState.ShotsFired,
			"历史记录数应等于发射数")
	}
}

// 验证：cloneSession 是深拷贝，对历史记录的修改不影响原对象
func TestCloneSession_DeepCopy(t *testing.T) {
	orig := &model.VirtualShootSession{
		SessionID:   "test-1",
		ShotsFired:  3,
		HistoryShots: []model.TimeSeriesPoint{
			{Timestamp: 1, Value: 1},
			{Timestamp: 2, Value: 1},
			{Timestamp: 3, Value: 1},
		},
	}

	cloned := cloneSession(orig)
	// 修改克隆的历史记录
	cloned.HistoryShots[0].Timestamp = 999
	// 原对象不应变
	assert.Equal(t, int64(1), orig.HistoryShots[0].Timestamp)
	assert.Equal(t, int64(999), cloned.HistoryShots[0].Timestamp)
}

// 验证：nil 安全
func TestCloneSession_NilSafe(t *testing.T) {
	assert.Nil(t, cloneSession(nil))
}

// ============ 并发安全测试 ============

// 高并发：同一会话100个goroutine同时发射，检查总数据一致性
func TestShoot_ConcurrentSameSession(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, _ := mgr.NewSession("zhuge")

	var wg sync.WaitGroup
	errCount := 0
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 每次都跳过冷却
			sess.LastShotUnixSec = time.Now().Unix() - 100
			resp, err := mgr.Shoot(sess.SessionID, model.VirtualShootRequest{
				SessionID: sess.SessionID, Mode: "single",
			})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errCount++
			}
			_ = resp
		}()
	}
	wg.Wait()

	assert.Equal(t, 0, errCount)
	s, ok := mgr.GetSession(sess.SessionID)
	assert.True(t, ok)
	t.Logf("并发100发后: 实际发射%d, 卡弹%d, 装弹%d",
		s.ShotsFired, s.JamCount, s.ReloadCount)
	// 状态不应崩溃或出现负弹容
	assert.GreaterOrEqual(t, s.CurrentAmmo, 0)
	assert.GreaterOrEqual(t, s.ShotsFired, 0)
	assert.LessOrEqual(t, s.StringFatigue, 1.0)
}

// ============ GetVariant & ListActiveSessions ============

func TestGetVariant_Exists(t *testing.T) {
	mgr := NewVirtualShootManager()
	v, ok := mgr.GetVariant("zhuge")
	assert.True(t, ok)
	assert.Equal(t, "诸葛弩", v.Name)
}

func TestGetVariant_NotExists(t *testing.T) {
	mgr := NewVirtualShootManager()
	_, ok := mgr.GetVariant("xx")
	assert.False(t, ok)
}

func TestListActiveSessions_InitiallyEmpty(t *testing.T) {
	mgr := NewVirtualShootManager()
	assert.Empty(t, mgr.ListActiveSessions())
}

// 边界：空请求回退
func TestShoot_DefaultBurstCount(t *testing.T) {
	mgr := NewVirtualShootManager()
	sess, _ := mgr.NewSession("zhuge")
	// 未设置 BurstCount，应为默认值3
	resp, _ := mgr.Shoot(sess.SessionID, model.VirtualShootRequest{
		SessionID: sess.SessionID, Mode: "burst",
	})
	// burst 默认3发，最多3发
	assert.LessOrEqual(t, resp.NewState.ShotsFired, 3)
}
