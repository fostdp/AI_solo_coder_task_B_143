import React, { useState, useEffect, useMemo } from 'react'
import {
  Card,
  Button,
  Checkbox,
  Table,
  Tag,
  Spin,
  Alert,
  Space,
  Typography,
  Row,
  Col,
  Statistic,
  Empty
} from 'antd'
import {
  HistoryOutlined,
  ThunderboltOutlined,
  AimOutlined,
  SwapOutlined
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { variantApi, firearmApi } from '../services/api'

const { Title, Text } = Typography

const MOCK_ANCIENT = [
  { code: 'zhuge', name: '诸葛弩', dynasty: '三国', description: '诸葛亮发明的连发弩', armLength: 85, stringTension: 450, magazineCapacity: 10, maxRange: 260, effectiveRange: 150, theoreticalFireRate: 12, reloadTime: 8, accuracyScore: 72, sustainabilityScore: 85 },
  { code: 'san-gong', name: '三弓弩', dynasty: '宋代', description: '三弓合力大型床弩', armLength: 220, stringTension: 1800, magazineCapacity: 1, maxRange: 1500, effectiveRange: 800, theoreticalFireRate: 0.5, reloadTime: 120, accuracyScore: 65, sustainabilityScore: 30 },
  { code: 'bi-zhang', name: '臂张弩', dynasty: '战国', description: '单兵臂力弩', armLength: 65, stringTension: 280, magazineCapacity: 1, maxRange: 180, effectiveRange: 100, theoreticalFireRate: 2, reloadTime: 15, accuracyScore: 88, sustainabilityScore: 60 }
]

const MOCK_MODERN = [
  { code: 'm1-garand', name: 'M1 加兰德', type: '半自动步枪', era: 1936, theoreticalFireRate: 50, actualFireRate: 45, magazineCapacity: 8, caliber: '7.62mm', effectiveRange: 457, muzzleVelocity: 865 },
  { code: 'ak-47', name: 'AK-47', type: '突击步枪', era: 1947, theoreticalFireRate: 600, actualFireRate: 100, magazineCapacity: 30, caliber: '7.62mm', effectiveRange: 350, muzzleVelocity: 715 },
  { code: 'm16a1', name: 'M16A1', type: '突击步枪', era: 1967, theoreticalFireRate: 800, actualFireRate: 55, magazineCapacity: 30, caliber: '5.56mm', effectiveRange: 500, muzzleVelocity: 990 },
  { code: 'mp5', name: 'HK MP5', type: '冲锋枪', era: 1966, theoreticalFireRate: 800, actualFireRate: 120, magazineCapacity: 30, caliber: '9mm', effectiveRange: 100, muzzleVelocity: 400 },
  { code: 'm249-saw', name: 'M249 SAW', type: '班用机枪', era: 1984, theoreticalFireRate: 900, actualFireRate: 200, magazineCapacity: 200, caliber: '5.56mm', effectiveRange: 800, muzzleVelocity: 915 },
  { code: 'desert-eagle', name: '沙漠之鹰', type: '大口径手枪', era: 1983, theoreticalFireRate: 30, actualFireRate: 20, magazineCapacity: 7, caliber: '12.7mm', effectiveRange: 50, muzzleVelocity: 470 }
]

const buildMockResult = (ancientCodes, modernCodes) => {
  const ancient = MOCK_ANCIENT.filter(v => ancientCodes.includes(v.code))
  const modern = MOCK_MODERN.filter(f => modernCodes.includes(f.code))
  const ancientAvgFireRate = ancient.reduce((s, v) => s + v.theoreticalFireRate, 0) / (ancient.length || 1)
  const modernAvgFireRate = modern.reduce((s, f) => s + f.actualFireRate, 0) / (modern.length || 1)
  const fireRateRatio = modernAvgFireRate / (ancientAvgFireRate || 1)
  const ancientAvgRange = ancient.reduce((s, v) => s + v.effectiveRange, 0) / (ancient.length || 1)
  const modernAvgRange = modern.reduce((s, f) => s + f.effectiveRange, 0) / (modern.length || 1)
  const rangeRatio = modernAvgRange / (ancientAvgRange || 1)
  const baseRate = ancient.length > 0 ? ancient.reduce((s, v) => s + v.theoreticalFireRate, 0) / ancient.length : 1
  const comparisonTable = [
    ...ancient.map(v => ({
      name: v.name, type: `${v.dynasty}弩`, era: v.dynasty === '战国' ? -300 : v.dynasty === '三国' ? 230 : 1100,
      theoreticalFireRate: v.theoreticalFireRate, actualFireRate: v.theoreticalFireRate,
      magazineCapacity: v.magazineCapacity, caliber: '箭矢', effectiveRange: v.effectiveRange,
      muzzleVelocity: 0, relativeRatio: v.theoreticalFireRate / (baseRate || 1)
    })),
    ...modern.map(f => ({
      name: f.name, type: f.type, era: f.era,
      theoreticalFireRate: f.theoreticalFireRate, actualFireRate: f.actualFireRate,
      magazineCapacity: f.magazineCapacity, caliber: f.caliber, effectiveRange: f.effectiveRange,
      muzzleVelocity: f.muzzleVelocity, relativeRatio: f.actualFireRate / (baseRate || 1)
    }))
  ]
  const dynastyYearMap = { '战国': -300, '三国': 230, '宋代': 1100 }
  const timeline = [
    ...ancient.map(v => ({ year: dynastyYearMap[v.dynasty] || 0, name: v.name, fireRate: v.theoreticalFireRate, type: 'ancient' })),
    ...modern.map(f => ({ year: f.era, name: f.name, fireRate: f.actualFireRate, type: 'modern' }))
  ]
  return { ancient, modern, stats: { ancientAvgFireRate, modernAvgFireRate, fireRateRatio, rangeRatio }, comparisonTable, timeline }
}

const EraFirepowerCompare = () => {
  const [ancientList, setAncientList] = useState(MOCK_ANCIENT)
  const [modernList, setModernList] = useState(MOCK_MODERN)
  const [selectedAncient, setSelectedAncient] = useState([])
  const [selectedModern, setSelectedModern] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const [result, setResult] = useState(null)
  const [initialLoading, setInitialLoading] = useState(true)

  useEffect(() => {
    const loadData = async () => {
      try {
        const variants = await variantApi.getVariants()
        if (variants && variants.length > 0) setAncientList(variants)
      } catch (err) {
        console.warn('使用默认弩型数据:', err.message)
      }
      try {
        const firearms = await firearmApi.getModernFirearms()
        if (firearms && firearms.length > 0) setModernList(firearms)
      } catch (err) {
        console.warn('使用默认火器数据:', err.message)
      }
      setInitialLoading(false)
    }
    loadData()
  }, [])

  const handleAncientChange = (code, checked) => {
    setSelectedAncient(prev => checked ? [...prev, code] : prev.filter(c => c !== code))
  }

  const handleModernChange = (code, checked) => {
    setSelectedModern(prev => checked ? [...prev, code] : prev.filter(c => c !== code))
  }

  const handleCompare = async () => {
    if (selectedAncient.length === 0 && selectedModern.length === 0) return
    setLoading(true)
    setError(null)
    try {
      const data = await firearmApi.compareEra({
        ancientVariants: selectedAncient,
        modernFirearms: selectedModern
      })
      setResult(data)
    } catch (err) {
      console.warn('使用模拟对比数据:', err.message)
      setResult(buildMockResult(selectedAncient, selectedModern))
    } finally {
      setLoading(false)
    }
  }

  const barChartOption = useMemo(() => {
    if (!result) return {}
    const ancientNames = result.ancient.map(v => v.name)
    const modernNames = result.modern.map(f => f.name)
    const allNames = [...ancientNames, ...modernNames]
    const ancientFireRates = result.ancient.map(v => v.theoreticalFireRate)
    const modernFireRates = result.modern.map(f => f.actualFireRate)
    const ancientRanges = result.ancient.map(v => v.effectiveRange)
    const modernRanges = result.modern.map(f => f.effectiveRange)
    const fireRateData = [
      ...ancientFireRates.map((v, i) => ({ value: v, itemStyle: { color: '#CD853F' } })),
      ...modernFireRates.map(() => null)
    ]
    const modernFireRateData = [
      ...ancientFireRates.map(() => null),
      ...modernFireRates.map(v => ({ value: v, itemStyle: { color: '#4682B4' } }))
    ]
    const rangeData = [
      ...ancientRanges.map((v) => ({ value: v, itemStyle: { color: '#CD853F' } })),
      ...modernRanges.map(() => null)
    ]
    const modernRangeData = [
      ...ancientRanges.map(() => null),
      ...modernRanges.map(v => ({ value: v, itemStyle: { color: '#4682B4' } }))
    ]
    return {
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' },
        textStyle: { color: '#d4c4a8' },
        backgroundColor: 'rgba(42, 31, 24, 0.95)',
        borderColor: '#8B4513'
      },
      legend: {
        data: ['古代弩-射速', '现代火器-射速', '古代弩-有效射程', '现代火器-有效射程'],
        textStyle: { color: '#d4c4a8' },
        bottom: 0
      },
      grid: { left: 80, right: 80, top: 40, bottom: 60 },
      xAxis: {
        type: 'category',
        data: allNames,
        axisLabel: { color: '#d4c4a8', rotate: 20 },
        axisLine: { lineStyle: { color: '#8B4513' } }
      },
      yAxis: [
        {
          type: 'log',
          name: '射速(发/分)',
          nameTextStyle: { color: '#d4c4a8' },
          axisLabel: { color: '#d4c4a8' },
          axisLine: { lineStyle: { color: '#8B4513' } },
          splitLine: { lineStyle: { color: 'rgba(139, 69, 19, 0.3)' } }
        },
        {
          type: 'value',
          name: '有效射程(m)',
          nameTextStyle: { color: '#d4c4a8' },
          axisLabel: { color: '#d4c4a8' },
          axisLine: { lineStyle: { color: '#4682B4' } },
          splitLine: { show: false }
        }
      ],
      series: [
        { name: '古代弩-射速', type: 'bar', yAxisIndex: 0, data: fireRateData, barWidth: 20 },
        { name: '现代火器-射速', type: 'bar', yAxisIndex: 0, data: modernFireRateData, barWidth: 20 },
        { name: '古代弩-有效射程', type: 'bar', yAxisIndex: 1, data: rangeData, barGap: '-100%', barWidth: 12, opacity: 0.6 },
        { name: '现代火器-有效射程', type: 'bar', yAxisIndex: 1, data: modernRangeData, barGap: '-100%', barWidth: 12, opacity: 0.6 }
      ]
    }
  }, [result])

  const timelineOption = useMemo(() => {
    if (!result) return {}
    const ancientData = result.timeline.filter(t => t.type === 'ancient').map(t => ({
      value: [t.year, t.fireRate, t.fireRate],
      name: t.name
    }))
    const modernData = result.timeline.filter(t => t.type === 'modern').map(t => ({
      value: [t.year, t.fireRate, t.fireRate],
      name: t.name
    }))
    return {
      tooltip: {
        trigger: 'item',
        formatter: (params) => {
          const d = params.data
          return `${d.name}<br/>年份: ${d.value[0]}<br/>射速: ${d.value[1]} 发/分`
        },
        textStyle: { color: '#d4c4a8' },
        backgroundColor: 'rgba(42, 31, 24, 0.95)',
        borderColor: '#8B4513'
      },
      legend: {
        data: ['古代弩', '现代火器'],
        textStyle: { color: '#d4c4a8' },
        bottom: 0
      },
      grid: { left: 80, right: 40, top: 40, bottom: 60 },
      xAxis: {
        type: 'value',
        name: '公元年份',
        min: -300,
        max: 2000,
        nameTextStyle: { color: '#d4c4a8' },
        axisLabel: { color: '#d4c4a8' },
        axisLine: { lineStyle: { color: '#8B4513' } },
        splitLine: { lineStyle: { color: 'rgba(139, 69, 19, 0.3)' } }
      },
      yAxis: {
        type: 'log',
        name: '射速(发/分)',
        nameTextStyle: { color: '#d4c4a8' },
        axisLabel: { color: '#d4c4a8' },
        axisLine: { lineStyle: { color: '#8B4513' } },
        splitLine: { lineStyle: { color: 'rgba(139, 69, 19, 0.3)' } }
      },
      series: [
        {
          name: '古代弩',
          type: 'scatter',
          data: ancientData,
          symbolSize: (val) => Math.max(Math.log(val[2]) * 8, 16),
          itemStyle: { color: '#B8860B', shadowBlur: 10, shadowColor: 'rgba(184, 134, 11, 0.5)' },
          label: {
            show: true,
            formatter: (p) => p.data.name,
            position: 'top',
            color: '#B8860B',
            fontSize: 12
          }
        },
        {
          name: '现代火器',
          type: 'scatter',
          data: modernData,
          symbolSize: (val) => Math.max(Math.log(val[2]) * 8, 16),
          itemStyle: { color: '#4682B4', shadowBlur: 10, shadowColor: 'rgba(70, 130, 180, 0.5)' },
          label: {
            show: true,
            formatter: (p) => p.data.name,
            position: 'top',
            color: '#4682B4',
            fontSize: 12
          }
        }
      ]
    }
  }, [result])

  const tableColumns = [
    { title: '名称', dataIndex: 'name', key: 'name', fixed: 'left', width: 120, render: (t, r) => <Text strong style={{ color: r.era < 0 || r.era < 1900 ? '#B8860B' : '#4682B4' }}>{t}</Text> },
    { title: '类型', dataIndex: 'type', key: 'type', width: 120, render: (t) => <Tag color="#5a4a3a">{t}</Tag> },
    { title: '年代', dataIndex: 'era', key: 'era', width: 80, align: 'center', render: (v) => <Text style={{ color: '#d4c4a8' }}>{v < 0 ? `公元前${Math.abs(v)}年` : `${v}年`}</Text> },
    { title: '理论射速', dataIndex: 'theoreticalFireRate', key: 'theoreticalFireRate', width: 110, align: 'center', render: (v) => <Text style={{ color: '#d4c4a8' }}>{v} 发/分</Text> },
    { title: '实际射速', dataIndex: 'actualFireRate', key: 'actualFireRate', width: 110, align: 'center', render: (v) => <Text style={{ color: '#d4c4a8' }}>{v} 发/分</Text> },
    { title: '弹容', dataIndex: 'magazineCapacity', key: 'magazineCapacity', width: 70, align: 'center', render: (v) => <Text style={{ color: '#d4c4a8' }}>{v}</Text> },
    { title: '口径', dataIndex: 'caliber', key: 'caliber', width: 80, align: 'center', render: (t) => <Tag color="#8B4513">{t}</Tag> },
    { title: '有效射程', dataIndex: 'effectiveRange', key: 'effectiveRange', width: 100, align: 'center', render: (v) => <Text style={{ color: '#d4c4a8' }}>{v} m</Text> },
    { title: '枪口初速', dataIndex: 'muzzleVelocity', key: 'muzzleVelocity', width: 100, align: 'center', render: (v) => <Text style={{ color: '#d4c4a8' }}>{v > 0 ? `${v} m/s` : '-'}</Text> },
    {
      title: '相对倍率', dataIndex: 'relativeRatio', key: 'relativeRatio', width: 100, align: 'center',
      render: (v) => <Tag color={v >= 1 ? '#2E8B57' : '#8B4513'}>{v.toFixed(1)}x</Tag>
    }
  ]

  if (initialLoading) {
    return (
      <div style={{ padding: 24, display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div style={{ padding: 24, overflow: 'auto', height: '100%' }}>
      <Card
        style={{
          background: 'linear-gradient(135deg, #2a1f18 0%, #1a1410 100%)',
          border: '1px solid #8B4513',
          marginBottom: 16
        }}
        bodyStyle={{ padding: 24 }}
      >
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div>
            <Title level={3} style={{ margin: 0, color: '#B8860B', fontFamily: "'Ma Shan Zheng', cursive" }}>
              <HistoryOutlined style={{ marginRight: 8 }} />
              跨时代火力对比
            </Title>
            <Text style={{ display: 'block', marginTop: 8, color: '#a89888', fontSize: 14 }}>
              千年兵器 VS 现代火器
            </Text>
          </div>

          <Row gutter={24}>
            <Col xs={24} md={12}>
              <Card
                size="small"
                title={
                  <Space>
                    <AimOutlined style={{ color: '#B8860B' }} />
                    <span style={{ color: '#d4c4a8' }}>古代弩</span>
                  </Space>
                }
                style={{ background: 'rgba(42, 31, 24, 0.6)', border: '1px solid #5a4a3a' }}
                headStyle={{ borderBottom: '1px solid #5a4a3a' }}
              >
                <Space direction="vertical" style={{ width: '100%' }}>
                  {ancientList.map(v => (
                    <Card
                      key={v.code}
                      size="small"
                      hoverable
                      style={{
                        background: selectedAncient.includes(v.code) ? 'rgba(184, 134, 11, 0.15)' : 'rgba(42, 31, 24, 0.4)',
                        border: selectedAncient.includes(v.code) ? '1px solid #B8860B' : '1px solid #5a4a3a',
                        cursor: 'pointer',
                        transition: 'all 0.3s'
                      }}
                      onClick={() => handleAncientChange(v.code, !selectedAncient.includes(v.code))}
                    >
                      <Space>
                        <Checkbox
                          checked={selectedAncient.includes(v.code)}
                          onChange={(e) => handleAncientChange(v.code, e.target.checked)}
                        />
                        <Text strong style={{ color: selectedAncient.includes(v.code) ? '#B8860B' : '#d4c4a8' }}>{v.name}</Text>
                        <Tag color="#8B4513">{v.dynasty}</Tag>
                      </Space>
                      <div style={{ marginTop: 4 }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>{v.description}</Text>
                      </div>
                    </Card>
                  ))}
                </Space>
              </Card>
            </Col>

            <Col xs={24} md={12}>
              <Card
                size="small"
                title={
                  <Space>
                    <ThunderboltOutlined style={{ color: '#4682B4' }} />
                    <span style={{ color: '#d4c4a8' }}>现代步枪</span>
                  </Space>
                }
                style={{ background: 'rgba(42, 31, 24, 0.6)', border: '1px solid #5a4a3a' }}
                headStyle={{ borderBottom: '1px solid #5a4a3a' }}
              >
                <Space direction="vertical" style={{ width: '100%' }}>
                  {modernList.map(f => (
                    <Card
                      key={f.code}
                      size="small"
                      hoverable
                      style={{
                        background: selectedModern.includes(f.code) ? 'rgba(70, 130, 180, 0.15)' : 'rgba(42, 31, 24, 0.4)',
                        border: selectedModern.includes(f.code) ? '1px solid #4682B4' : '1px solid #5a4a3a',
                        cursor: 'pointer',
                        transition: 'all 0.3s'
                      }}
                      onClick={() => handleModernChange(f.code, !selectedModern.includes(f.code))}
                    >
                      <Space>
                        <Checkbox
                          checked={selectedModern.includes(f.code)}
                          onChange={(e) => handleModernChange(f.code, e.target.checked)}
                        />
                        <Text strong style={{ color: selectedModern.includes(f.code) ? '#4682B4' : '#d4c4a8' }}>{f.name}</Text>
                        <Tag color="#2E8B57">{f.type}</Tag>
                      </Space>
                      <div style={{ marginTop: 4 }}>
                        <Text type="secondary" style={{ fontSize: 12 }}>{f.era}年 | {f.caliber} | 射速 {f.actualFireRate} 发/分</Text>
                      </div>
                    </Card>
                  ))}
                </Space>
              </Card>
            </Col>
          </Row>

          <Space>
            <Button
              type="primary"
              size="large"
              icon={<SwapOutlined />}
              onClick={handleCompare}
              disabled={(selectedAncient.length === 0 && selectedModern.length === 0) || loading}
              loading={loading}
              style={{
                background: 'linear-gradient(135deg, #B8860B 0%, #8B4513 100%)',
                border: 'none',
                minWidth: 180
              }}
            >
              跨时代对比
            </Button>
            <Text type="secondary">
              已选 古代{selectedAncient.length} / 现代{selectedModern.length}
            </Text>
          </Space>
        </Space>
      </Card>

      {error && (
        <Alert
          type="error"
          message={error}
          showIcon
          closable
          style={{ marginBottom: 16 }}
          onClose={() => setError(null)}
        />
      )}

      {loading && (
        <Card style={{ background: '#2a1f18', border: '1px solid #8B4513' }}>
          <div style={{ textAlign: 'center', padding: 60 }}>
            <Spin size="large" />
            <div style={{ marginTop: 16, color: '#a89888' }}>正在进行跨时代火力分析...</div>
          </div>
        </Card>
      )}

      {result && !loading && (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Row gutter={16}>
            <Col xs={12} sm={6}>
              <Card style={{ background: 'linear-gradient(135deg, rgba(184, 134, 11, 0.2) 0%, rgba(139, 69, 19, 0.2) 100%)', border: '1px solid #B8860B', textAlign: 'center' }}>
                <Statistic
                  title={<span style={{ color: '#a89888' }}>古代平均射速</span>}
                  value={result.stats.ancientAvgFireRate}
                  suffix="发/分"
                  valueStyle={{ color: '#B8860B' }}
                  prefix={<AimOutlined />}
                />
              </Card>
            </Col>
            <Col xs={12} sm={6}>
              <Card style={{ background: 'linear-gradient(135deg, rgba(70, 130, 180, 0.2) 0%, rgba(42, 31, 24, 0.3) 100%)', border: '1px solid #4682B4', textAlign: 'center' }}>
                <Statistic
                  title={<span style={{ color: '#a89888' }}>现代平均射速</span>}
                  value={result.stats.modernAvgFireRate}
                  suffix="发/分"
                  valueStyle={{ color: '#4682B4' }}
                  prefix={<ThunderboltOutlined />}
                />
              </Card>
            </Col>
            <Col xs={12} sm={6}>
              <Card style={{ background: 'linear-gradient(135deg, rgba(46, 139, 87, 0.2) 0%, rgba(42, 31, 24, 0.3) 100%)', border: '1px solid #2E8B57', textAlign: 'center' }}>
                <Statistic
                  title={<span style={{ color: '#a89888' }}>射速倍数差</span>}
                  value={result.stats.fireRateRatio}
                  suffix="x"
                  precision={1}
                  valueStyle={{ color: '#2E8B57' }}
                  prefix={<SwapOutlined />}
                />
              </Card>
            </Col>
            <Col xs={12} sm={6}>
              <Card style={{ background: 'linear-gradient(135deg, rgba(205, 133, 63, 0.2) 0%, rgba(42, 31, 24, 0.3) 100%)', border: '1px solid #CD853F', textAlign: 'center' }}>
                <Statistic
                  title={<span style={{ color: '#a89888' }}>射程倍数差</span>}
                  value={result.stats.rangeRatio}
                  suffix="x"
                  precision={1}
                  valueStyle={{ color: '#CD853F' }}
                  prefix={<HistoryOutlined />}
                />
              </Card>
            </Col>
          </Row>

          <Card
            title={
              <Space>
                <SwapOutlined style={{ color: '#B8860B' }} />
                <span style={{ color: '#d4c4a8' }}>射速与射程双轴对比</span>
              </Space>
            }
            style={{ background: '#2a1f18', border: '1px solid #8B4513' }}
            headStyle={{ borderBottom: '1px solid #8B4513' }}
          >
            <div style={{ height: 420 }}>
              <ReactECharts
                option={barChartOption}
                style={{ height: '100%', width: '100%' }}
                notMerge
                lazyUpdate
                opts={{ renderer: 'canvas' }}
              />
            </div>
          </Card>

          <Card
            title={
              <Space>
                <AimOutlined style={{ color: '#B8860B' }} />
                <span style={{ color: '#d4c4a8' }}>跨时代火力对比表</span>
              </Space>
            }
            style={{ background: '#2a1f18', border: '1px solid #8B4513' }}
            headStyle={{ borderBottom: '1px solid #8B4513' }}
          >
            <Table
              dataSource={result.comparisonTable.map((r, i) => ({ ...r, key: i }))}
              columns={tableColumns}
              pagination={false}
              size="middle"
              scroll={{ x: 'max-content' }}
              style={{ background: 'transparent' }}
            />
          </Card>

          <Card
            title={
              <Space>
                <HistoryOutlined style={{ color: '#B8860B' }} />
                <span style={{ color: '#d4c4a8' }}>历史射速演进时间轴</span>
              </Space>
            }
            style={{ background: '#2a1f18', border: '1px solid #8B4513' }}
            headStyle={{ borderBottom: '1px solid #8B4513' }}
          >
            <div style={{ height: 420 }}>
              <ReactECharts
                option={timelineOption}
                style={{ height: '100%', width: '100%' }}
                notMerge
                lazyUpdate
                opts={{ renderer: 'canvas' }}
              />
            </div>
          </Card>
        </Space>
      )}

      {!result && !loading && !error && (
        <Card style={{ background: '#2a1f18', border: '1px solid #8B4513' }}>
          <Empty
            description={<span style={{ color: '#a89888' }}>请选择古代弩和现代火器后点击「跨时代对比」</span>}
          />
        </Card>
      )}
    </div>
  )
}

export default EraFirepowerCompare
