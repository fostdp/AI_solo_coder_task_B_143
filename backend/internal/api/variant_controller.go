package api

import (
	"net/http"
	"strconv"

	"crossbow-simulation/backend/internal/coordinator"
	"crossbow-simulation/backend/internal/model"
	"crossbow-simulation/backend/internal/reliability"
	"crossbow-simulation/backend/internal/repository"
	"crossbow-simulation/backend/internal/virtual_shoot"
	"crossbow-simulation/backend/internal/mechanism_comparator"
	"crossbow-simulation/backend/internal/era_comparator"
	"crossbow-simulation/backend/internal/feed_reliability_analyzer"
	"crossbow-simulation/backend/internal/vr_crossbow"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Controller struct {
	repo      *repository.Repository
	coord     *coordinator.Coordinator
	analyzer  *reliability.MagazineReliabilityAnalyzer
	shootMgr  *virtual_shoot.VirtualShootManager
	upgrader  websocket.Upgrader
	// new modular services (feature 4 services below:
	mechComp *mechanism_comparator.Comparator
	eraComp  *era_comparator.Comparator
	feedAnalyzerFactory func(variant *model.CrossbowVariant) *feed_reliability_analyzer.Analyzer
	vrShootMgr *vr_crossbow.SessionManager
}

func NewFeatureController() *Controller {
	defaultVariant := model.CrossbowPresets()[0]
	defaultParams := reliability.BuildParamsFromVariant(&defaultVariant)
	return &Controller{
		repo: nil, coord: nil,
		analyzer: reliability.NewMagazineReliabilityAnalyzer(defaultParams),
		shootMgr: virtual_shoot.NewVirtualShootManager(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		mechComp: mechanism_comparator.NewComparator(),
		eraComp:  era_comparator.NewComparator(),
		feedAnalyzerFactory: func(v *model.CrossbowVariant) *feed_reliability_analyzer.Analyzer {
			params := feed_reliability_analyzer.BuildParamsFromVariant(v)
			return feed_reliability_analyzer.NewAnalyzer(params)
		},
		vrShootMgr: vr_crossbow.NewSessionManager(),
	}
}

// ================= 弩型机构对比 =================

func (ctrl *Controller) ListVariants(c *gin.Context) {
	items := ctrl.mechComp.GetAll()
	c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: map[string]interface{}{"items": items, "total": len(items)}})
}

func (ctrl *Controller) GetVariant(c *gin.Context) {
	code := c.Param("code")
	v, ok := ctrl.mechComp.GetByCode(code)
	if !ok {
		c.JSON(http.StatusNotFound, model.APIResponse{Success: false, Message: "弩型未找到"})
		return
	}
	c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: v})
}

func (ctrl *Controller) CompareVariants(c *gin.Context) {
	var req model.VariantCompareRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.VariantCodes) == 0 {
		req.VariantCodes = mechanism_comparator.DefaultVariantCodes()
	}
	if len(req.CompareMetrics) == 0 {
		req.CompareMetrics = mechanism_comparator.DefaultMetrics()
	}
	result, err := ctrl.mechComp.Compare(mechanism_comparator.CompareOptions{
		VariantCodes:   req.VariantCodes,
		CompareMetrics: req.CompareMetrics,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: err.Error()})
		return
	}
	if len(result.ComparedVariants) == 0 {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: "无有效弩型编码"})
		return
	}
	resp := model.VariantCompareResponse{
		ComparedVariants: result.ComparedVariants,
		PerformanceRadar: result.PerformanceRadar,
		AdvantageMap:     result.AdvantageMap,
	}
	c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: resp, Warnings: result.Errors})
}

// ================= 跨时代对比 =================

func (ctrl *Controller) ListModernFirearms(c *gin.Context) {
	items := ctrl.eraComp.ListModern()
	c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: map[string]interface{}{"items": items, "total": len(items)}})
}

func (ctrl *Controller) CompareEraFirearms(c *gin.Context) {
	var req model.EraFirearmCompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = model.EraFirearmCompareRequest{
			AncientVariants: era_comparator.DefaultAncientCodes(),
			ModernFirearms:  era_comparator.DefaultModernNames(),
		}
	}
	if len(req.AncientVariants) == 0 { req.AncientVariants = era_comparator.DefaultAncientCodes() }
	if len(req.ModernFirearms) == 0 { req.ModernFirearms = era_comparator.DefaultModernNames() }
	if len(req.CompareMetrics) == 0 { req.CompareMetrics = era_comparator.DefaultMetrics() }

	result := ctrl.eraComp.Compare(era_comparator.CompareOptions{
		AncientCodes: req.AncientVariants,
		ModernNames:  req.ModernFirearms,
		Metrics:      req.CompareMetrics,
	})
	resp := model.EraFirearmCompareResponse{
		AncientVariants: result.AncientVariants,
		ModernFirearms:  result.ModernFirearms,
		EraGapTable:     result.EraGapTable,
	}
	c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: resp, Warnings: result.Errors})
}

// ================= 供弹可靠性分析 =================

func (ctrl *Controller) AnalyzeMagazineReliability(c *gin.Context) {
	code := c.Param("code")
	v, ok := ctrl.mechComp.GetByCode(code)
	if !ok {
		c.JSON(http.StatusNotFound, model.APIResponse{Success: false, Message: "弩型未找到"})
		return
	}
	var body struct {
		Shots      int     `json:"shots"`
		SimTimeSec float64 `json:"simTimeSec"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		body.Shots, _ = strconv.Atoi(c.DefaultQuery("shots", "5000"))
		st, _ := strconv.ParseFloat(c.DefaultQuery("simTime", "7200"), 64)
		body.SimTimeSec = st
	}
	if body.Shots <= 0 { body.Shots = 5000 }
	if body.SimTimeSec <= 0 { body.SimTimeSec = 7200 }

	analyzer := ctrl.feedAnalyzerFactory(&v)
	analysis := analyzer.Analyze(body.Shots, body.SimTimeSec)
	report := analyzer.GenerateReport(code, analysis)

	c.JSON(http.StatusOK, model.APIResponse{Success: true,
		Data: map[string]interface{}{"variant": v, "analysis": analysis, "report": report}})
}

// ================= 虚拟射击体验 =================

func (ctrl *Controller) ensureVR() {
	if ctrl.vrShootMgr == nil {
		ctrl.vrShootMgr = vr_crossbow.NewSessionManager()
	}
	if ctrl.mechComp == nil {
		ctrl.mechComp = mechanism_comparator.NewComparator()
	}
	if ctrl.eraComp == nil {
		ctrl.eraComp = era_comparator.NewComparator()
	}
	if ctrl.feedAnalyzerFactory == nil {
		ctrl.feedAnalyzerFactory = func(v *model.CrossbowVariant) *feed_reliability_analyzer.Analyzer {
			params := feed_reliability_analyzer.BuildParamsFromVariant(v)
			return feed_reliability_analyzer.NewAnalyzer(params)
		}
	}
}

func (ctrl *Controller) StartVirtualShoot(c *gin.Context) {
	ctrl.ensureVR()
	var req struct { VariantCode string `json:"variantCode"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		req.VariantCode = c.DefaultQuery("variant", "zhuge")
	}
	if req.VariantCode == "" { req.VariantCode = "zhuge" }
	sess, err := ctrl.vrShootMgr.NewSession(req.VariantCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Success: false, Message: "创建射击会话失败"})
		return
	}
	c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: sess})
}

func (ctrl *Controller) ShootAction(c *gin.Context) {
	ctrl.ensureVR()
	var req model.VirtualShootRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: "无效的请求参数"})
		return
	}
	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: "缺少 sessionId"})
		return
	}
	resp, err := ctrl.vrShootMgr.Shoot(req.SessionID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{Success: false, Message: "射击失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: resp})
}

func (ctrl *Controller) GetVirtualShootStatus(c *gin.Context) {
	ctrl.ensureVR()
	sid := c.Param("id")
	sess, ok := ctrl.vrShootMgr.GetSession(sid)
	if !ok {
		c.JSON(http.StatusNotFound, model.APIResponse{Success: false, Message: "会话不存在"})
		return
	}
	c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: sess})
}

func (ctrl *Controller) ResetVirtualShoot(c *gin.Context) {
	ctrl.ensureVR()
	sid := c.Param("id")
	sess, ok := ctrl.vrShootMgr.ResetSession(sid)
	if !ok {
		c.JSON(http.StatusNotFound, model.APIResponse{Success: false, Message: "会话不存在"})
		return
	}
	stats := ctrl.vrShootMgr.EngineStats()
	c.JSON(http.StatusOK, model.APIResponse{Success: true, Message: "会话已重置",
		Data: map[string]interface{}{"session": sess, "engineStats": stats}})
}
