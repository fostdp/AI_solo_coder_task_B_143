import React, { useState, useEffect, useRef, useCallback } from 'react'
import {
  Card,
  Button,
  Select,
  Progress,
  Spin,
  Alert,
  Space,
  Typography,
  Row,
  Col,
  Statistic,
  Tag,
  Empty,
  Tooltip
} from 'antd'
import {
  PlayCircleOutlined,
  ReloadOutlined,
  AimOutlined,
  ThunderboltOutlined,
  WarningOutlined,
  FireOutlined,
  SafetyCertificateOutlined,
  ClockCircleOutlined,
  BoxPlotOutlined,
  DashboardOutlined
} from '@ant-design/icons'
import { variantApi, virtualShootApi } from '../services/api'

const { Title, Text } = Typography
const { Option } = Select

const MOCK_VARIANTS = [
  { code: 'zhuge', name: '诸葛弩', dynasty: '三国', accuracyScore: 72, magazineCapacity: 10, theoreticalFireRate: 12 },
  { code: 'san-gong', name: '三弓弩', dynasty: '宋代', accuracyScore: 65, magazineCapacity: 1, theoreticalFireRate: 0.5 },
  { code: 'bi-zhang', name: '臂张弩', dynasty: '战国', accuracyScore: 88, magazineCapacity: 1, theoreticalFireRate: 2 }
]

const ACCURACY_DEVIATION_MAP = {
  'zhuge': 18,
  'san-gong': 22,
  'bi-zhang': 12
}

const VARIANT_COLORS = {
  'zhuge': '#B8860B',
  'san-gong': '#CD5C5C',
  'bi-zhang': '#4682B4'
}

const gaussianRandom = (mean, std) => {
  const u1 = Math.random()
  const u2 = Math.random()
  const z = Math.sqrt(-2 * Math.log(u1)) * Math.cos(2 * Math.PI * u2)
  return mean + z * std
}

const VirtualShootingGallery = () => {
  const [variants, setVariants] = useState(MOCK_VARIANTS)
  const [selectedCode, setSelectedCode] = useState('zhuge')
  const [sessionID, setSessionID] = useState(null)
  const [started, setStarted] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [jamAlert, setJamAlert] = useState(null)
  const [shake, setShake] = useState(false)

  const [currentAmmo, setCurrentAmmo] = useState(10)
  const [magazineCapacity, setMagazineCapacity] = useState(10)
  const [totalShots, setTotalShots] = useState(0)
  const [jams, setJams] = useState(0)
  const [stringFatigue, setStringFatigue] = useState(0)
  const [reloading, setReloading] = useState(false)
  const [reloadProgress, setReloadProgress] = useState(0)
  const [cooling, setCooling] = useState(false)
  const [coolTimeRemaining, setCoolTimeRemaining] = useState(0)

  const [hitMarks, setHitMarks] = useState([])
  const [popupRing, setPopupRing] = useState(null)
  const [arrowFlying, setArrowFlying] = useState(false)
  const [arrowProgress, setArrowProgress] = useState(0)
  const [arrowTarget, setArrowTarget] = useState({ x: 0, y: 0, ring: 0 })

  const [recentShotTimes, setRecentShotTimes] = useState([])
  const sessionStartRef = useRef(null)
  const canvasRef = useRef(null)
  const animationRef = useRef(null)
  const reloadTimerRef = useRef(null)
  const coolTimerRef = useRef(null)
  const pollTimerRef = useRef(null)

  useEffect(() => {
    const loadVariants = async () => {
      try {
        const data = await variantApi.getVariants()
        if (data && data.length > 0) setVariants(data)
      } catch (err) {
        console.warn('使用默认弩型数据:', err.message)
      }
    }
    loadVariants()
  }, [])

  const selectedVariant = variants.find(v => v.code === selectedCode) || MOCK_VARIANTS[0]

  const calcInstantRPM = useCallback(() => {
    const now = Date.now()
    const windowStart = now - 10000
    const recent = recentShotTimes.filter(t => t >= windowStart)
    return recent.length * 6
  }, [recentShotTimes])

  const calcAvgRPM = useCallback(() => {
    if (!sessionStartRef.current || totalShots === 0) return 0
    const elapsedMin = (Date.now() - sessionStartRef.current) / 60000
    if (elapsedMin < 1 / 60) return 0
    return totalShots / elapsedMin
  }, [totalShots])

  const drawCanvas = useCallback(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    const W = canvas.width
    const H = canvas.height
    const targetX = W - 160
    const targetY = H / 2
    const targetR = 130

    ctx.clearRect(0, 0, W, H)

    const bgGrad = ctx.createLinearGradient(0, 0, W, H)
    bgGrad.addColorStop(0, 'rgba(42, 31, 24, 0.9)')
    bgGrad.addColorStop(1, 'rgba(26, 20, 16, 0.95)')
    ctx.fillStyle = bgGrad
    ctx.fillRect(0, 0, W, H)

    ctx.strokeStyle = 'rgba(139, 69, 19, 0.2)'
    ctx.lineWidth = 1
    for (let x = 0; x < W; x += 40) {
      ctx.beginPath()
      ctx.moveTo(x, 0)
      ctx.lineTo(x, H)
      ctx.stroke()
    }
    for (let y = 0; y < H; y += 40) {
      ctx.beginPath()
      ctx.moveTo(0, y)
      ctx.lineTo(W, y)
      ctx.stroke()
    }

    const bowX = 120
    const bowY = H / 2
    const bowColor = VARIANT_COLORS[selectedCode] || '#B8860B'

    ctx.save()
    ctx.translate(bowX, bowY)

    ctx.fillStyle = '#5a3a20'
    ctx.strokeStyle = '#8B4513'
    ctx.lineWidth = 3
    ctx.beginPath()
    ctx.roundRect(-8, -5, 100, 10, 4)
    ctx.fill()
    ctx.stroke()

    ctx.strokeStyle = bowColor
    ctx.lineWidth = 6
    ctx.lineCap = 'round'
    ctx.beginPath()
    ctx.arc(15, 0, 70, -Math.PI / 2.2, Math.PI / 2.2)
    ctx.stroke()

    ctx.strokeStyle = '#f5deb3'
    ctx.lineWidth = 2
    ctx.setLineDash([5, 3])
    ctx.beginPath()
    const stringTensionOffset = arrowFlying ? -20 + arrowProgress * 40 : 0
    ctx.moveTo(15, -70)
    ctx.quadraticCurveTo(30 + stringTensionOffset, 0, 15, 70)
    ctx.stroke()
    ctx.setLineDash([])

    ctx.fillStyle = bowColor
    ctx.strokeStyle = '#8B4513'
    ctx.lineWidth = 2
    ctx.beginPath()
    ctx.arc(5, 0, 14, 0, Math.PI * 2)
    ctx.fill()
    ctx.stroke()

    if (!arrowFlying && currentAmmo > 0 && !reloading && !cooling) {
      ctx.fillStyle = '#CD853F'
      ctx.strokeStyle = '#8B4513'
      ctx.lineWidth = 1
      ctx.beginPath()
      ctx.moveTo(95, 0)
      ctx.lineTo(55, -3)
      ctx.lineTo(55, 3)
      ctx.closePath()
      ctx.fill()
      ctx.stroke()

      ctx.strokeStyle = '#fff'
      ctx.lineWidth = 1.5
      ctx.beginPath()
      ctx.moveTo(55, 0)
      ctx.lineTo(45, -5)
      ctx.moveTo(55, 0)
      ctx.lineTo(45, 5)
      ctx.stroke()
    }

    ctx.restore()

    ctx.strokeStyle = 'rgba(139, 69, 19, 0.4)'
    ctx.lineWidth = 2
    ctx.setLineDash([10, 5])
    ctx.beginPath()
    ctx.moveTo(bowX + 100, bowY)
    ctx.lineTo(targetX - targetR, targetY)
    ctx.stroke()
    ctx.setLineDash([])

    for (let i = 10; i >= 1; i--) {
      const r = (targetR / 10) * i
      const isEven = i % 2 === 0
      ctx.beginPath()
      ctx.arc(targetX, targetY, r, 0, Math.PI * 2)
      if (i === 10) {
        ctx.fillStyle = '#f5deb3'
      } else if (i >= 9) {
        ctx.fillStyle = '#FFD700'
      } else if (i >= 7) {
        ctx.fillStyle = isEven ? '#CD5C5C' : '#f5deb3'
      } else if (i >= 5) {
        ctx.fillStyle = isEven ? '#4682B4' : '#f5deb3'
      } else {
        ctx.fillStyle = isEven ? '#2E8B57' : '#f5deb3'
      }
      ctx.fill()
      ctx.strokeStyle = '#8B4513'
      ctx.lineWidth = 1.5
      ctx.stroke()
    }

    ctx.fillStyle = '#1a1410'
    ctx.font = 'bold 14px serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    for (let i = 1; i <= 10; i++) {
      const labelR = (targetR / 10) * (i - 0.5)
      if (i === 10) {
        ctx.fillText('X', targetX, targetY)
      } else {
        ctx.fillText(String(i), targetX + labelR - 6, targetY - 2)
      }
    }

    ctx.strokeStyle = '#8B4513'
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.moveTo(targetX, targetY - targetR - 10)
    ctx.lineTo(targetX, targetY + targetR + 10)
    ctx.moveTo(targetX - targetR - 10, targetY)
    ctx.lineTo(targetX + targetR + 10, targetY)
    ctx.stroke()

    hitMarks.forEach((hit, idx) => {
      const rad = (targetR / 10) * 10
      const px = targetX + (hit.offsetX / 100) * rad
      const py = targetY + (hit.offsetY / 100) * rad
      ctx.save()
      ctx.translate(px, py)
      ctx.rotate(Math.random() * 0.5 - 0.25)
      ctx.strokeStyle = hit.ring >= 9 ? '#FFD700' : '#2E8B57'
      ctx.lineWidth = 2
      ctx.beginPath()
      ctx.moveTo(-8, 0)
      ctx.lineTo(8, 0)
      ctx.moveTo(0, -8)
      ctx.lineTo(0, 8)
      ctx.stroke()
      ctx.beginPath()
      ctx.arc(0, 0, 5, 0, Math.PI * 2)
      ctx.stroke()
      ctx.restore()
    })

    if (arrowFlying) {
      const ax = bowX + 95 + (targetX - (bowX + 95) - 50) * arrowProgress
      const ay = bowY + (arrowTarget.y - bowY) * arrowProgress
      const angle = Math.atan2(arrowTarget.y - bowY, (targetX - bowX - 50))
      ctx.save()
      ctx.translate(ax, ay)
      ctx.rotate(angle)
      ctx.fillStyle = '#CD853F'
      ctx.strokeStyle = '#8B4513'
      ctx.lineWidth = 1
      ctx.beginPath()
      ctx.moveTo(30, 0)
      ctx.lineTo(-10, -4)
      ctx.lineTo(-10, 4)
      ctx.closePath()
      ctx.fill()
      ctx.stroke()
      ctx.strokeStyle = '#fff'
      ctx.lineWidth = 1.5
      ctx.beginPath()
      ctx.moveTo(-10, 0)
      ctx.lineTo(-18, -6)
      ctx.moveTo(-10, 0)
      ctx.lineTo(-18, 6)
      ctx.stroke()
      ctx.restore()
    }

    ctx.fillStyle = 'rgba(184, 134, 11, 0.9)'
    ctx.font = 'bold 16px "Ma Shan Zheng", serif'
    ctx.textAlign = 'left'
    ctx.textBaseline = 'top'
    ctx.fillText(`${selectedVariant.name} 虚拟射击场`, 20, 20)
    ctx.font = '12px serif'
    ctx.fillStyle = '#a89888'
    ctx.fillText(`精度评分: ${selectedVariant.accuracyScore} | 弹容: ${magazineCapacity}发`, 20, 45)

    ctx.textAlign = 'right'
    ctx.fillStyle = '#d4c4a8'
    ctx.font = '12px serif'
    ctx.fillText(`共命中: ${hitMarks.length} 发`, W - 20, 20)
    if (hitMarks.length > 0) {
      const avgRing = hitMarks.reduce((s, h) => s + h.ring, 0) / hitMarks.length
      ctx.fillText(`平均环数: ${avgRing.toFixed(1)}环`, W - 20, 40)
    }
  }, [selectedCode, selectedVariant, hitMarks, arrowFlying, arrowProgress, arrowTarget, currentAmmo, magazineCapacity, reloading, cooling])

  useEffect(() => {
    drawCanvas()
  }, [drawCanvas])

  useEffect(() => {
    const loop = () => {
      drawCanvas()
      animationRef.current = requestAnimationFrame(loop)
    }
    animationRef.current = requestAnimationFrame(loop)
    return () => {
      if (animationRef.current) cancelAnimationFrame(animationRef.current)
    }
  }, [drawCanvas])

  useEffect(() => {
    const onKey = (e) => {
      if (e.code === 'Space' && started && !loading) {
        e.preventDefault()
        handleShoot()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [started, loading])

  const cleanupTimers = () => {
    if (reloadTimerRef.current) clearInterval(reloadTimerRef.current)
    if (coolTimerRef.current) clearInterval(coolTimerRef.current)
    if (pollTimerRef.current) clearInterval(pollTimerRef.current)
    reloadTimerRef.current = null
    coolTimerRef.current = null
    pollTimerRef.current = null
  }

  useEffect(() => {
    return cleanupTimers
  }, [])

  const handleStart = async () => {
    setLoading(true)
    setError(null)
    cleanupTimers()
    try {
      let resp
      try {
        resp = await virtualShootApi.start(selectedCode)
      } catch (err) {
        console.warn('使用模拟 start:', err.message)
        const v = MOCK_VARIANTS.find(x => x.code === selectedCode)
        resp = {
          sessionID: `sess_${Date.now()}`,
          variantCode: selectedCode,
          magazineCapacity: v?.magazineCapacity || 10,
          currentAmmo: v?.magazineCapacity || 10
        }
      }
      setSessionID(resp.sessionID)
      setCurrentAmmo(resp.currentAmmo)
      setMagazineCapacity(resp.magazineCapacity)
      setTotalShots(0)
      setJams(0)
      setStringFatigue(0)
      setReloading(false)
      setReloadProgress(0)
      setCooling(false)
      setCoolTimeRemaining(0)
      setHitMarks([])
      setRecentShotTimes([])
      sessionStartRef.current = Date.now()
      setStarted(true)
    } catch (err) {
      setError(err.message || '启动失败')
    } finally {
      setLoading(false)
    }
  }

  const handleReset = async () => {
    setLoading(true)
    setError(null)
    cleanupTimers()
    try {
      if (sessionID) {
        try {
          await virtualShootApi.reset(sessionID)
        } catch (err) {
          console.warn('reset error (ignored):', err.message)
        }
      }
    } catch (ignored) {}
    setSessionID(null)
    setStarted(false)
    setCurrentAmmo(magazineCapacity)
    setTotalShots(0)
    setJams(0)
    setStringFatigue(0)
    setReloading(false)
    setReloadProgress(0)
    setCooling(false)
    setCoolTimeRemaining(0)
    setHitMarks([])
    setRecentShotTimes([])
    setJamAlert(null)
    setLoading(false)
  }

  const startReload = () => {
    setReloading(true)
    setReloadProgress(0)
    const variant = MOCK_VARIANTS.find(v => v.code === selectedCode) || MOCK_VARIANTS[0]
    const reloadTimeMs = variant.code === 'san-gong' ? 8000 : variant.code === 'bi-zhang' ? 2500 : 1500
    const steps = 30
    let step = 0
    reloadTimerRef.current = setInterval(() => {
      step++
      setReloadProgress(Math.round((step / steps) * 100))
      if (step >= steps) {
        clearInterval(reloadTimerRef.current)
        reloadTimerRef.current = null
        setReloading(false)
        setReloadProgress(100)
        setCurrentAmmo(magazineCapacity)
        setTimeout(() => setReloadProgress(0), 300)
      }
    }, reloadTimeMs / steps)
  }

  const simulateHit = () => {
    const deviation = ACCURACY_DEVIATION_MAP[selectedCode] || 15
    const ox = gaussianRandom(0, deviation)
    const oy = gaussianRandom(0, deviation)
    const dist = Math.sqrt(ox * ox + oy * oy)
    let ring
    if (dist <= 3) ring = 10
    else if (dist <= 10) ring = 9
    else if (dist <= 20) ring = 8
    else if (dist <= 30) ring = 7
    else if (dist <= 40) ring = 6
    else if (dist <= 50) ring = 5
    else if (dist <= 60) ring = 4
    else if (dist <= 70) ring = 3
    else if (dist <= 85) ring = 2
    else ring = 1
    return { offsetX: ox, offsetY: oy, ring }
  }

  const animateArrow = (hit) => {
    setArrowTarget(hit)
    setArrowFlying(true)
    setArrowProgress(0)
    let t = 0
    const duration = 250
    const startTime = Date.now()
    const anim = () => {
      const elapsed = Date.now() - startTime
      const p = Math.min(1, elapsed / duration)
      const eased = 1 - Math.pow(1 - p, 3)
      setArrowProgress(eased)
      if (eased < 1) {
        requestAnimationFrame(anim)
      } else {
        setArrowFlying(false)
        setHitMarks(prev => [...prev.slice(-40), hit])
        setPopupRing({ ring: hit.ring, id: Date.now() })
        setTimeout(() => setPopupRing(null), 1000)
      }
    }
    requestAnimationFrame(anim)
  }

  const handleShoot = useCallback(async () => {
    if (!started || reloading || cooling || arrowFlying || loading) return
    if (currentAmmo <= 0) {
      startReload()
      return
    }
    setLoading(true)
    try {
      let resp
      try {
        resp = await virtualShootApi.shoot(sessionID)
      } catch (err) {
        console.warn('使用模拟 shoot:', err.message)
        const jamChance = selectedCode === 'zhuge' ? 0.008 : selectedCode === 'san-gong' ? 0.002 : 0.005
        const jammed = Math.random() < jamChance
        const hit = simulateHit()
        resp = {
          shotFired: !jammed,
          jammed,
          recovered: jammed,
          recoverTime: jammed ? 1.2 + Math.random() * 1.5 : 0,
          newState: {
            currentAmmo: Math.max(0, currentAmmo - 1),
            totalShots: totalShots + (jammed ? 0 : 1),
            jams: jams + (jammed ? 1 : 0),
            stringFatigue: Math.min(100, stringFatigue + 0.3 + Math.random() * 0.5),
            reloading: currentAmmo - 1 <= 0,
            reloadProgress: 0,
            cooling: false,
            coolTimeRemaining: 0
          },
          hit: jammed ? undefined : hit
        }
      }

      if (resp.jammed) {
        setJams(j => j + 1)
        setShake(true)
        setJamAlert({
          message: `⚠ 机构卡弹！已自动排障耗时 ${resp.recoverTime?.toFixed?.(1) || '1.2'} 秒`,
          id: Date.now()
        })
        setTimeout(() => setShake(false), 500)
        setTimeout(() => setJamAlert(null), 3500)
      }

      if (resp.shotFired) {
        const now = Date.now()
        setRecentShotTimes(prev => [...prev.filter(t => t >= now - 10000), now])
        setTotalShots(t => t + 1)
        setCurrentAmmo(resp.newState.currentAmmo)
        const hit = resp.hit || simulateHit()
        animateArrow(hit)
      }

      setStringFatigue(resp.newState.stringFatigue)

      if (resp.newState.currentAmmo <= 0 && !reloading) {
        setTimeout(() => startReload(), 300)
      }

      if (resp.newState.cooling || resp.newState.stringFatigue >= 95) {
        setCooling(true)
        setCoolTimeRemaining(resp.newState.coolTimeRemaining || 2000)
        coolTimerRef.current = setInterval(() => {
          setCoolTimeRemaining(prev => {
            const next = prev - 100
            if (next <= 0) {
              clearInterval(coolTimerRef.current)
              coolTimerRef.current = null
              setCooling(false)
              return 0
            }
            return next
          })
        }, 100)
      }
    } catch (err) {
      setError(err.message || '发射失败')
    } finally {
      setLoading(false)
    }
  }, [started, reloading, cooling, arrowFlying, loading, currentAmmo, sessionID, selectedCode, totalShots, jams, stringFatigue])

  const instantRPM = calcInstantRPM()
  const avgRPM = calcAvgRPM()

  const canShoot = started && !reloading && !cooling && !arrowFlying && !loading && currentAmmo > 0

  return (
    <div
      style={{
        padding: 24,
        overflow: 'auto',
        height: '100%',
        animation: shake ? 'shake 0.4s cubic-bezier(.36,.07,.19,.97) both' : 'none'
      }}
    >
      <style>{`
        @keyframes shake {
          10%, 90% { transform: translate3d(-1px, 0, 0); }
          20%, 80% { transform: translate3d(2px, 0, 0); }
          30%, 50%, 70% { transform: translate3d(-4px, 0, 0); }
          40%, 60% { transform: translate3d(4px, 0, 0); }
        }
        @keyframes ringPop {
          0% { transform: scale(0.3); opacity: 0; }
          30% { transform: scale(1.3); opacity: 1; }
          100% { transform: scale(1); opacity: 0; }
        }
        .ring-popup {
          animation: ringPop 1s ease-out forwards;
        }
      `}</style>

      <Card
        style={{
          background: 'linear-gradient(135deg, #2a1f18 0%, #1a1410 100%)',
          border: '1px solid #8B4513',
          marginBottom: 16
        }}
        bodyStyle={{ padding: '20px 24px' }}
      >
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div style={{ textAlign: 'center' }}>
            <Title level={2} style={{ margin: 0, color: '#B8860B', fontFamily: "'Ma Shan Zheng', cursive", letterSpacing: 4 }}>
              <AimOutlined style={{ marginRight: 12 }} />
              古代连弩虚拟射击体验馆
              <AimOutlined style={{ marginLeft: 12 }} />
            </Title>
            <Text type="secondary" style={{ fontSize: 15, display: 'block', marginTop: 6 }}>
              感受古人的连发火力 · 体验千年兵器的魅力
            </Text>
          </div>

          <Row gutter={24} align="middle" justify="center">
            <Col>
              <Space>
                <Text type="secondary">弩型选择：</Text>
                <Select
                  value={selectedCode}
                  onChange={(c) => {
                    setSelectedCode(c)
                    if (started) handleReset()
                    const v = MOCK_VARIANTS.find(x => x.code === c)
                    if (v) {
                      setMagazineCapacity(v.magazineCapacity)
                      setCurrentAmmo(v.magazineCapacity)
                    }
                  }}
                  style={{ width: 200 }}
                  size="large"
                  disabled={started}
                >
                  {variants.map(v => (
                    <Option key={v.code} value={v.code}>
                      <Space>
                        <span style={{ color: VARIANT_COLORS[v.code] }}>●</span>
                        <Text strong>{v.name}</Text>
                        <Tag color="#8B4513">{v.dynasty}</Tag>
                      </Space>
                    </Option>
                  ))}
                </Select>
              </Space>
            </Col>
            <Col>
              <Space>
                <Button
                  type={started ? 'default' : 'primary'}
                  size="large"
                  icon={started ? <ReloadOutlined /> : <PlayCircleOutlined />}
                  onClick={started ? handleReset : handleStart}
                  loading={loading}
                  style={{
                    background: started
                      ? 'linear-gradient(135deg, #3a2f28 0%, #2a1f18 100%)'
                      : 'linear-gradient(135deg, #2E8B57 0%, #1a5a37 100%)',
                    border: started ? '1px solid #8B4513' : 'none',
                    color: started ? '#d4c4a8' : '#fff',
                    minWidth: 130
                  }}
                >
                  {started ? '重置体验' : '开始体验'}
                </Button>
                {started && (
                  <Tag color="#2E8B57" icon={<SafetyCertificateOutlined />}>
                    会话: {sessionID?.slice(-8) || 'active'}
                  </Tag>
                )}
              </Space>
            </Col>
          </Row>
        </Space>
      </Card>

      {error && (
        <Alert type="error" message={error} showIcon closable style={{ marginBottom: 16 }} onClose={() => setError(null)} />
      )}
      {jamAlert && (
        <Alert
          type="warning"
          message={jamAlert.message}
          showIcon
          closable
          style={{ marginBottom: 16, border: '1px solid #CD5C5C', background: 'rgba(205, 92, 92, 0.15)' }}
          onClose={() => setJamAlert(null)}
        />
      )}

      <Row gutter={16}>
        <Col xs={24} xl={16}>
          <Card
            style={{
              background: '#2a1f18',
              border: '1px solid #8B4513',
              position: 'relative',
              overflow: 'hidden'
            }}
            bodyStyle={{ padding: 0 }}
          >
            <div style={{ position: 'relative' }}>
              <canvas
                ref={canvasRef}
                width={900}
                height={500}
                style={{ width: '100%', height: 'auto', display: 'block', cursor: canShoot ? 'crosshair' : 'not-allowed' }}
                onClick={canShoot ? handleShoot : undefined}
              />

              {popupRing && (
                <div
                  key={popupRing.id}
                  className="ring-popup"
                  style={{
                    position: 'absolute',
                    right: 180,
                    top: '40%',
                    fontSize: 48,
                    fontWeight: 'bold',
                    fontFamily: "'Ma Shan Zheng', cursive",
                    color: popupRing.ring >= 9 ? '#FFD700' : popupRing.ring >= 7 ? '#CD5C5C' : popupRing.ring >= 5 ? '#4682B4' : '#2E8B57',
                    textShadow: '0 0 20px currentColor, 0 2px 8px rgba(0,0,0,0.5)',
                    pointerEvents: 'none',
                    zIndex: 10
                  }}
                >
                  {popupRing.ring === 10 ? '🎯 X!' : `${popupRing.ring}环`}
                </div>
              )}

              {!started && (
                <div style={{
                  position: 'absolute',
                  inset: 0,
                  background: 'rgba(26, 20, 16, 0.75)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  backdropFilter: 'blur(3px)'
                }}>
                  <Empty
                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                    description={
                      <Space direction="vertical" size={8} style={{ alignItems: 'center' }}>
                        <Text strong style={{ color: '#B8860B', fontSize: 16 }}>
                          点击「开始体验」进入虚拟射击
                        </Text>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          发射方式：空格键 / 点击发射按钮 / 点击箭靶
                        </Text>
                      </Space>
                    }
                  />
                </div>
              )}
            </div>

            <div style={{
              padding: '16px 24px',
              borderTop: '1px solid #8B4513',
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              gap: 16,
              background: 'rgba(0,0,0,0.2)'
            }}>
              <Tooltip title={canShoot ? '点击发射 (空格)' : reloading ? '装弹中...' : cooling ? '冷却中...' : '开始后可用'}>
                <Button
                  type="primary"
                  size="large"
                  icon={<FireOutlined />}
                  onClick={handleShoot}
                  disabled={!canShoot}
                  loading={loading}
                  style={{
                    background: canShoot
                      ? 'linear-gradient(135deg, #CD5C5C 0%, #8B0000 100%)'
                      : undefined,
                    border: 'none',
                    minWidth: 200,
                    minHeight: 56,
                    fontSize: 20,
                    fontWeight: 'bold',
                    fontFamily: "'Ma Shan Zheng', cursive",
                    letterSpacing: 4,
                    boxShadow: canShoot ? '0 8px 24px rgba(205, 92, 92, 0.4)' : 'none'
                  }}
                >
                  🏹 发射！
                </Button>
              </Tooltip>

              {cooling && (
                <Tag color="orange" icon={<ClockCircleOutlined />}>
                  冷却剩余: {(coolTimeRemaining / 1000).toFixed(1)}s
                </Tag>
              )}
            </div>
          </Card>
        </Col>

        <Col xs={24} xl={8}>
          <Card
            title={
              <Space>
                <DashboardOutlined style={{ color: '#B8860B' }} />
                <span style={{ color: '#d4c4a8' }}>射击HUD 状态面板</span>
              </Space>
            }
            style={{ background: '#2a1f18', border: '1px solid #8B4513', height: '100%' }}
            headStyle={{ borderBottom: '1px solid #8B4513' }}
            bodyStyle={{ padding: 16 }}
          >
            <Space direction="vertical" size={14} style={{ width: '100%' }}>
              <div>
                <Space justify="space-between" style={{ width: '100%', marginBottom: 6 }}>
                  <Space>
                    <BoxPlotOutlined style={{ color: '#B8860B' }} />
                    <Text type="secondary">当前弹容</Text>
                  </Space>
                  <Text strong style={{ color: '#d4c4a8' }}>
                    {currentAmmo} / {magazineCapacity}
                  </Text>
                </Space>
                <Progress
                  percent={Math.round((currentAmmo / magazineCapacity) * 100)}
                  showInfo={false}
                  strokeColor={
                    currentAmmo === 0 ? '#CD5C5C'
                      : currentAmmo <= magazineCapacity * 0.3 ? '#CD853F'
                      : '#2E8B57'
                  }
                  trailColor="#1a1410"
                  size="small"
                />
                {reloading && (
                  <div style={{ marginTop: 8 }}>
                    <Space size={4}>
                      <Text type="secondary" style={{ fontSize: 12 }}>🔄 装弹中</Text>
                      <Tag color="orange">{reloadProgress}%</Tag>
                    </Space>
                    <Progress
                      percent={reloadProgress}
                      showInfo={false}
                      status="active"
                      strokeColor="#B8860B"
                      trailColor="#1a1410"
                      size="small"
                      style={{ marginTop: 4 }}
                    />
                  </div>
                )}
              </div>

              <Row gutter={12}>
                <Col span={12}>
                  <Card size="small" style={{ background: 'rgba(70, 130, 180, 0.1)', border: '1px solid #4682B4' }}>
                    <Statistic
                      title={<Text type="secondary" style={{ color: '#4682B4', fontSize: 11 }}>已发射总数</Text>}
                      value={totalShots}
                      valueStyle={{ color: '#4682B4', fontSize: 20 }}
                      suffix="发"
                    />
                  </Card>
                </Col>
                <Col span={12}>
                  <Card size="small" style={{ background: 'rgba(205, 92, 92, 0.1)', border: '1px solid #CD5C5C' }}>
                    <Statistic
                      title={<Text type="secondary" style={{ color: '#CD5C5C', fontSize: 11 }}>卡弹次数</Text>}
                      value={jams}
                      valueStyle={{ color: '#CD5C5C', fontSize: 20 }}
                      suffix="次"
                    />
                  </Card>
                </Col>
              </Row>

              <Row gutter={12}>
                <Col span={12}>
                  <Card size="small" style={{ background: 'rgba(46, 139, 87, 0.1)', border: '1px solid #2E8B57' }}>
                    <Tooltip title="最近10秒发射数 × 6">
                      <Statistic
                        title={<Text type="secondary" style={{ color: '#2E8B57', fontSize: 11 }}>瞬时射速</Text>}
                        value={instantRPM}
                        precision={0}
                        valueStyle={{ color: '#2E8B57', fontSize: 18 }}
                        suffix="RPM"
                      />
                    </Tooltip>
                  </Card>
                </Col>
                <Col span={12}>
                  <Card size="small" style={{ background: 'rgba(184, 134, 11, 0.1)', border: '1px solid #B8860B' }}>
                    <Statistic
                      title={<Text type="secondary" style={{ color: '#B8860B', fontSize: 11 }}>平均射速</Text>}
                      value={avgRPM}
                      precision={1}
                      valueStyle={{ color: '#B8860B', fontSize: 18 }}
                      suffix="RPM"
                    />
                  </Card>
                </Col>
              </Row>

              <div>
                <Space justify="space-between" style={{ width: '100%', marginBottom: 6 }}>
                  <Space>
                    <ThunderboltOutlined style={{
                      color: stringFatigue >= 80 ? '#CD5C5C' : stringFatigue >= 50 ? '#CD853F' : '#2E8B57'
                    }} />
                    <Text type="secondary">弓弦疲劳度</Text>
                  </Space>
                  <Text strong style={{
                    color: stringFatigue >= 80 ? '#CD5C5C' : stringFatigue >= 50 ? '#CD853F' : '#d4c4a8'
                  }}>
                    {stringFatigue.toFixed(1)}%
                  </Text>
                </Space>
                <Progress
                  percent={Math.min(100, stringFatigue)}
                  showInfo={false}
                  strokeColor={
                    stringFatigue >= 80 ? { '0%': '#CD5C5C', '100%': '#8B0000' }
                      : stringFatigue >= 50 ? { '0%': '#CD853F', '100%': '#B8860B' }
                      : { '0%': '#2E8B57', '100%': '#1a5a37' }
                  }
                  status={stringFatigue >= 95 ? 'exception' : undefined}
                  trailColor="#1a1410"
                />
              </div>

              <Divider style={{ margin: '4px 0', borderColor: '#5a4a3a' }} />

              <Card size="small" style={{ background: 'rgba(42, 31, 24, 0.5)', border: '1px solid #5a4a3a' }}>
                <Space direction="vertical" size={4} style={{ width: '100%' }}>
                  <Text strong style={{ color: '#B8860B', fontSize: 13 }}>
                    📊 本次命中统计
                  </Text>
                  {hitMarks.length === 0 ? (
                    <Text type="secondary" style={{ fontSize: 12 }}>尚未命中任何箭靶</Text>
                  ) : (
                    <>
                      <Row gutter={8}>
                        <Col span={12}>
                          <Text type="secondary" style={{ fontSize: 12 }}>命中总数：</Text>
                          <Text strong style={{ color: '#2E8B57' }}>{hitMarks.length} 发</Text>
                        </Col>
                        <Col span={12}>
                          <Text type="secondary" style={{ fontSize: 12 }}>平均环数：</Text>
                          <Text strong style={{ color: '#FFD700' }}>
                            {(hitMarks.reduce((s, h) => s + h.ring, 0) / hitMarks.length).toFixed(1)} 环
                          </Text>
                        </Col>
                      </Row>
                      <div style={{ display: 'flex', gap: 3, flexWrap: 'wrap', marginTop: 6 }}>
                        {[10, 9, 8, 7, 6, 5].map(r => {
                          const cnt = hitMarks.filter(h => h.ring === r).length
                          if (cnt === 0) return null
                          return (
                            <Tag
                              key={r}
                              color={r >= 9 ? 'gold' : r >= 7 ? 'red' : r >= 5 ? 'blue' : 'default'}
                              style={{ fontSize: 11, margin: 0 }}
                            >
                              {r === 10 ? 'X' : r}环: {cnt}
                            </Tag>
                          )
                        })}
                      </div>
                    </>
                  )}
                </Space>
              </Card>

              <Card size="small" style={{ background: 'rgba(139, 69, 19, 0.05)', border: '1px dashed #8B4513' }}>
                <Space direction="vertical" size={2}>
                  <Text type="secondary" style={{ fontSize: 11, color: '#B8860B' }}>
                    💡 操作提示
                  </Text>
                  <Text type="secondary" style={{ fontSize: 11 }}>
                    · 按 <Tag color="blue" style={{ fontSize: 10 }}>空格键</Tag> 快速发射
                  </Text>
                  <Text type="secondary" style={{ fontSize: 11 }}>
                    · 直接点击箭靶区域也可发射
                  </Text>
                  <Text type="secondary" style={{ fontSize: 11 }}>
                    · 弹夹射空后自动装弹
                  </Text>
                </Space>
              </Card>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default VirtualShootingGallery
