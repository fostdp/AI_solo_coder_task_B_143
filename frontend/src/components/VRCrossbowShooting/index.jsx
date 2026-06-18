import React, { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { Card, Button, Row, Col, Statistic, Progress, Slider, Radio, Space, Tag, Alert, Typography, Divider, Result } from 'antd';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ScatterChart, Scatter, ZAxis } from 'recharts';
import { variantApi } from '../../services/api';

const { Text, Title } = Typography;

const VARIANT_OPTIONS = [
  { code: 'zhuge', name: '诸葛弩', fireRate: 10.5, capacity: 10, triggerForceN: 38 },
  { code: 'san-gong', name: '三弓弩', fireRate: 1.5, capacity: 1, triggerForceN: 45 },
  { code: 'bi-zhang', name: '臂张弩', fireRate: 4, capacity: 1, triggerForceN: 28 },
];

const HAND_OPTIONS = [
  { key: 'right', label: '右手' },
  { key: 'left', label: '左手' },
  { key: 'both', label: '双手' },
];

const VRCrossbowShooting = ({ variantCode, onSessionReady }) => {
  const [selectedVariant, setSelectedVariant] = useState(variantCode || 'zhuge');
  const [handed, setHanded] = useState('right');
  const [sessionId, setSessionId] = useState(null);
  const [session, setSession] = useState(null);
  const [shotsLog, setShotsLog] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [triggerValue, setTriggerValue] = useState(0);
  const [triggerHistory, setTriggerHistory] = useState([]);
  const [firing, setFiring] = useState(false);
  const [coolingPercent, setCoolingPercent] = useState(100);
  const [latencyMs, setLatencyMs] = useState(0);
  const [engineStats, setEngineStats] = useState(null);
  const [impulseHistory, setImpulseHistory] = useState([]);
  const [vibrationHistory, setVibrationHistory] = useState([]);
  const [reloading, setReloading] = useState(false);
  const [hapticHint, setHapticHint] = useState(null);

  const fireIntervalRef = useRef(null);
  const lastShotsCount = useRef(0);

  const variant = VARIANT_OPTIONS.find(v => v.code === selectedVariant);

  const startSession = async (vc = selectedVariant) => {
    setLoading(true);
    setError(null);
    try {
      const res = await variantApi.startVirtualShoot({ variantCode: vc });
      setSession(res.data.session || res.data);
      setSessionId((res.data.session || res.data).sessionId);
      setShotsLog([]);
      setTriggerHistory([]);
      setImpulseHistory([]);
      setVibrationHistory([]);
      lastShotsCount.current = 0;
      if (onSessionReady) onSessionReady(res.data);
    } catch (e) {
      setError(e?.message || '创建会话失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { startSession(selectedVariant); }, [selectedVariant]);

  const doShoot = useCallback(async (triggerTravel = 100, pullSpeed = 0.3) => {
    if (!sessionId || reloading) return;
    const start = performance.now();
    try {
      const res = await variantApi.shootAction({
        sessionId, triggerTravelPct: triggerTravel, triggerPullSpeedMps: pullSpeed, operatorHanded: handed
      });
      const dt = performance.now() - start;
      setLatencyMs(dt);
      const resp = res.data;
      const s = res.data.session || resp;
      setSession(s);
      if (resp.jammed) {
        setHapticHint('卡弹！卡住了');
        setTimeout(() => setHapticHint(null), 1500);
      }
      if (resp.triggerFeedback) {
        setTriggerHistory(resp.triggerFeedback.forceCurve || []);
      }
      const impulse = resp.muzzleImpulseNs || (resp.shotResult?.muzzleImpulseNs);
      const vib = resp.bowVibrationHz || (resp.shotResult?.bowVibrationHz);
      if (impulse) setImpulseHistory(h => [...h.slice(-49), { i: h.length, v: impulse }]);
      if (vib) setVibrationHistory(h => [...h.slice(-49), { i: h.length, v: vib }]);
      setShotsLog(l => [...l.slice(-19), {
        n: l.length + 1,
        ok: !resp.jammed,
        cooling: resp.coolingTempC,
        rps: resp.rpm,
        latency: resp.releaseLatencyMs || 0,
        jam: resp.jammed,
        message: resp.message
      }]);
      if (s?.shotsFired && s.shotsFired > lastShotsCount.current) {
        setCoolingPercent(Math.max(0, 100 - (s.coolingTempC / 60) * 100));
        lastShotsCount.current = s.shotsFired;
      }
      if (resp.message?.includes('过热')) {
        setHapticHint('过热！请冷却');
      }
      if (resp.message?.includes('空仓')) {
        setHapticHint('空仓，自动装填中...');
        setReloading(true);
        setTimeout(() => setReloading(false), 2000);
      }
    } catch (e) {
      setError(e?.message || '射击失败');
    }
  }, [sessionId, handed, reloading]);

  const handleMouseDown = () => {
    if (reloading) return;
    setFiring(true);
    let pct = 0;
    let speed = 0;
    fireIntervalRef.current = setInterval(() => {
      if (pct < 100) {
        pct = Math.min(100, pct + 25);
        speed = 0.2 + (pct / 100) * 0.3;
        setTriggerValue(pct);
      } else {
        clearInterval(fireIntervalRef.current);
        fireIntervalRef.current = null;
        doShoot(100, speed);
        setTimeout(() => setTriggerValue(0), 150);
      }
    }, 40);
  };

  const handleMouseUp = () => {
    if (fireIntervalRef.current) {
      clearInterval(fireIntervalRef.current);
      fireIntervalRef.current = null;
    }
    if (triggerValue > 50) {
      doShoot(triggerValue, 0.3);
    }
    setTriggerValue(0);
    setFiring(false);
  };

  const handleReset = async () => {
    if (!sessionId) return;
    try {
      const res = await variantApi.resetVirtualShoot(sessionId);
      setSession(res.data.session);
      setEngineStats(res.data.engineStats);
      setShotsLog([]);
      setTriggerHistory([]);
      setImpulseHistory([]);
      setVibrationHistory([]);
      lastShotsCount.current = 0;
      setCoolingPercent(100);
      setHapticHint('已重置');
      setTimeout(() => setHapticHint(null), 1000);
    } catch (e) {
      setError(e?.message || '重置失败');
    }
  };

  const effectiveFireRate = useMemo(() => {
    if (!session?.rpm) return 0;
    const tenSec = shotsLog.slice(-Math.round(10 * variant.fireRate / 60));
    const jams = tenSec.filter(s => s.jam).length;
    const adjust = 1 - jams / Math.max(1, tenSec.length);
    return session.rpm * adjust;
  }, [session, shotsLog, variant]);

  const triggerChart = triggerHistory.map((p, i) => ({ x: i, 力: p.forceN }));

  const impulseChart = impulseHistory.slice(-20);
  const vibChart = vibrationHistory.slice(-20);

  return (
    <div className="vr-crossbow-shooting" style={{ padding: 16 }}>
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={6}>
          <Card title="射手参数" size="small" loading={loading} extra={<Button type="primary" onClick={handleReset}>重置</Button>}>
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <Text strong>弩型：</Text>
                <Radio.Group value={selectedVariant} onChange={e => { setSelectedVariant(e.target.value); }}>
                  {VARIANT_OPTIONS.map(v => <Radio.Button key={v.code} value={v.code}>{v.name}</Radio.Button>)}
                </Radio.Group>
              </div>
              <div>
                <Text strong>用手：</Text>
                <Radio.Group value={handed} onChange={e => setHanded(e.target.value)}>
                  {HAND_OPTIONS.map(h => <Radio.Button key={h.key} value={h.key}>{h.label}</Radio.Button>)}
                </Radio.Group>
              </div>
              <div>
                <Title level={5} style={{ margin: '12px 0 4px' }}>扳机</Title>
                <div
                  onMouseDown={handleMouseDown}
                  onMouseUp={handleMouseUp}
                  onMouseLeave={handleMouseUp}
                  style={{
                    position: 'relative',
                    width: '100%', height: 80,
                    background: `linear-gradient(to right, #f0f0f0 0%, #f0f0f0 ${triggerValue}%, #fff ${triggerValue}%, #fff 100%)`,
                    border: '2px solid #1890ff', borderRadius: 8,
                    cursor: firing ? 'grabbing' : 'grab',
                    userSelect: 'none', textAlign: 'center', lineHeight: '76px',
                    fontSize: 18, fontWeight: 600, color: '#1890ff',
                    transition: 'background 0.05s'
                  }}
                >
                  {firing ? `击发中 ${triggerValue}%` : (triggerValue > 0 ? `预压 ${triggerValue}%` : '按住扳机击发')}
                </div>
                <Text type="secondary" style={{fontSize:11}}>按住扳机逐渐下压 → 达到80%以上击发</Text>
              </div>
              <Button block type="danger" onClick={() => doShoot(100, 0.4)} disabled={reloading}>
                快速击发（Space）
              </Button>
              {hapticHint && <Alert type="warning" message={hapticHint} showIcon />}
              {error && <Alert type="error" message={error} />}
            </Space>
          </Card>
        </Col>
        <Col xs={24} lg={18}>
          <Row gutter={[8, 8]}>
            <Col xs={8} md={4}><Card size="small"><Statistic title="总发射" value={session?.shotsFired || 0} /></Card></Col>
            <Col xs={8} md={4}><Card size="small"><Statistic title="当前射速" value={effectiveFireRate} precision={1} suffix="RPM" valueStyle={{ color: '#52c41a' }} /></Card></Col>
            <Col xs={8} md={4}><Card size="small"><Statistic title="卡弹次数" value={session?.jams || 0} valueStyle={{ color: '#f5222d' }} /></Card></Col>
            <Col xs={8} md={4}><Card size="small"><Statistic title="冷却" value={coolingPercent} suffix="%" /></Card></Col>
            <Col xs={8} md={4}><Card size="small"><Statistic title="响应延迟" value={latencyMs} precision={0} suffix="ms" /></Card></Col>
            <Col xs={8} md={4}><Card size="small"><Statistic title="弹容" value={session?.remainingInMag || 0} suffix={`/${variant?.capacity}`} /></Card></Col>
          </Row>
          <Row gutter={[8, 8]} style={{ marginTop: 8 }}>
            <Col xs={24} md={12}>
              <Card title="扳机力-行程曲线（N vs 21采样点）" size="small"
                extra={session?.triggerProfileName ? <Tag color="blue">{session.triggerProfileName}</Tag> : null}>
                <ResponsiveContainer width="100%" height={180}>
                  <LineChart data={triggerChart}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="x" />
                    <YAxis />
                    <Tooltip />
                    <Line type="monotone" dataKey="力" stroke="#1890ff" dot={false} strokeWidth={2} />
                  </LineChart>
                </ResponsiveContainer>
              </Card>
            </Col>
            <Col xs={24} md={6}>
              <Card title="枪口冲量 (N·s)" size="small">
                <ResponsiveContainer width="100%" height={180}>
                  <LineChart data={impulseChart}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="i" />
                    <YAxis domain={['auto', 'auto']} />
                    <Line type="monotone" dataKey="v" stroke="#722ed1" strokeWidth={2} dot={false} />
                  </LineChart>
                </ResponsiveContainer>
              </Card>
            </Col>
            <Col xs={24} md={6}>
              <Card title="弓体振动 (Hz)" size="small">
                <ResponsiveContainer width="100%" height={180}>
                  <LineChart data={vibChart}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="i" />
                    <YAxis domain={['auto', 'auto']} />
                    <Line type="monotone" dataKey="v" stroke="#eb2f96" strokeWidth={2} dot={false} />
                  </LineChart>
                </ResponsiveContainer>
              </Card>
            </Col>
          </Row>
          <Row gutter={[8, 8]} style={{ marginTop: 8 }}>
            <Col xs={24} md={14}>
              <Card title="冷却温度 (°C)" size="small">
                <Progress percent={coolingPercent} format={(p) => `${(60 * (1 - p / 100)).toFixed(1)}°C`} />
                <div style={{height:60, marginTop:8}}>
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart data={shotsLog.slice(-30)}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="n" />
                      <YAxis domain={[0, 60]} />
                      <Line type="monotone" dataKey="cooling" stroke="#faad14" strokeWidth={2} dot={false} />
                    </LineChart>
                  </ResponsiveContainer>
                </div>
              </Card>
            </Col>
            <Col xs={24} md={10}>
              <Card title="最近发射记录" size="small" bodyStyle={{maxHeight:160, overflow:'auto'}}>
                {shotsLog.length === 0 ? <Text type="secondary">尚未发射</Text> : (
                  shotsLog.slice(-10).reverse().map((s, i) => (
                    <div key={i} style={{display:'flex', justifyContent:'space-between', padding:'2px 0', fontSize:12}}>
                      <span>#{s.n}: {s.ok ? <Tag color="green">成功</Tag> : <Tag color="red">卡弹</Tag>}</span>
                      <span style={{color:'#888'}}>{s.rps?.toFixed(0)} RPM · {s.latency}ms</span>
                    </div>
                  ))
                )}
              </Card>
            </Col>
          </Row>
        </Col>
        {engineStats && (
          <Col span={24}>
            <Alert type="info" showIcon
              message={`动力学引擎状态：已处理${engineStats.totalSubmitted} 任务 / 完成${engineStats.completed} / 拒绝${engineStats.rejected} / 平均延迟${engineStats.avgLatencyMs?.toFixed(1)||0}ms`} />
          </Col>
        )}
        {(!sessionId) && <Col span={24}><Result status="warning" title="请先开始一个会话" extra={<Button type="primary" onClick={() => startSession()}>开始</Button>} /></Col>}
      </Row>
    </div>
  );
};

export default VRCrossbowShooting;
