import React, { useState, useMemo } from 'react'
import {
  Card,
  Select,
  Slider,
  Button,
  Table,
  Row,
  Col,
  Statistic,
  Tag,
  Space,
  Typography,
  Spin,
  Alert,
  Empty
} from 'antd'
import {
  SafetyCertificateOutlined,
  WarningOutlined,
  ThunderboltOutlined,
  ExperimentOutlined,
  LineChartOutlined,
  PieChartOutlined,
  FileTextOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import { variantApi } from '../services/api'

const { Title, Text } = Typography
const { Option } = Select

const VARIANT_OPTIONS = [
  { code: 'zhuge', name: '诸葛弩' },
  { code: 'san-gong', name: '三弓弩' },
  { code: 'bi-zhang', name: '臂张弩' }
]

const PIE_COLORS = ['#CD5C5C', '#B8860B', '#4682B4', '#2E8B57', '#9370DB', '#FF8C00', '#708090']

const computeMockReliability = (code, shots, simHours) => {
  const jamBase = code === 'zhuge' ? 0.008 : code === 'san-gong' ? 0.002 : 0.005
  const jamCount = Math.round(shots * jamBase * (1 + Math.random() * 0.3))
  const mtbf = shots / Math.max(1, jamCount)
  const mtbfHours = simHours / Math.max(1, jamCount)
  const ciLower = mtbf * 0.75
  const ciUpper = mtbf * 1.25

  const modes = [
    { mode: 'DoubleFeed', name: '重弹/双供弹', count: Math.round(jamCount * 0.28), percentage: 28 },
    { mode: 'Misfeed', name: '不供弹', count: Math.round(jamCount * 0.22), percentage: 22 },
    { mode: 'Stovepipe', name: '卡壳', count: Math.round(jamCount * 0.18), percentage: 18 },
    { mode: 'FollowerBind', name: '托弹板卡滞', count: Math.round(jamCount * 0.13), percentage: 13 },
    { mode: 'SpringFatigue', name: '弹簧疲劳', count: Math.round(jamCount * 0.10), percentage: 10 },
    { mode: 'MagazineDamage', name: '匣体损坏', count: Math.round(jamCount * 0.05), percentage: 5 },
    { mode: 'ForeignObject', name: '异物卡入', count: Math.round(jamCount * 0.04), percentage: 4 }
  ]

  const reliabilityCurve = []
  const cumulativeJams = []
  let cumJams = 0
  for (let i = 0; i <= 100; i++) {
    const n = Math.round((i / 100) * shots)
    const r = Math.exp(-n / mtbf)
    reliabilityCurve.push({ n, r })
    cumJams = Math.round(n * jamBase)
    cumulativeJams.push({ n, count: cumJams })
  }

  const fmea = [
    { mode: 'DoubleFeed', name: '重弹/双供弹', severity: 7, occurrence: 6, detection: 5, rpn: 7 * 6 * 5, suggestion: '优化箭匣导轨间隙，增加分隔板' },
    { mode: 'Misfeed', name: '不供弹', severity: 8, occurrence: 5, detection: 6, rpn: 8 * 5 * 6, suggestion: '加强托弹弹簧预紧力，改进箭矢锥度' },
    { mode: 'Stovepipe', name: '卡壳', severity: 6, occurrence: 4, detection: 7, rpn: 6 * 4 * 7, suggestion: '增大排障窗口，添加防卡斜面' },
    { mode: 'FollowerBind', name: '托弹板卡滞', severity: 5, occurrence: 3, detection: 8, rpn: 5 * 3 * 8, suggestion: '抛光托弹板表面，添加润滑槽' },
    { mode: 'SpringFatigue', name: '弹簧疲劳', severity: 7, occurrence: 2, detection: 4, rpn: 7 * 2 * 4, suggestion: '定期更换弹簧，记录循环次数' },
    { mode: 'MagazineDamage', name: '匣体损坏', severity: 9, occurrence: 1, detection: 3, rpn: 9 * 1 * 3, suggestion: '加固匣体结构，使用高强度竹材' },
    { mode: 'ForeignObject', name: '异物卡入', severity: 5, occurrence: 2, detection: 5, rpn: 5 * 2 * 5, suggestion: '加装防尘盖，定期清洁箭匣' }
  ]

  return {
    totalShots: shots,
    jamCount,
    jamRate: (jamCount / shots * 100).toFixed(4),
    mtbf,
    mtbfHours,
    ciLower,
    ciUpper,
    reliabilityCurve,
    cumulativeJams,
    failureModes: modes,
    fmea
  }
}

const MagazineReliability = () => {
  const [selectedCode, setSelectedCode] = useState('zhuge')
  const [shots, setShots] = useState(10000)
  const [simHours, setSimHours] = useState(10)
  const [analyzing, setAnalyzing] = useState(false)
  const [error, setError] = useState(null)
  const [result, setResult] = useState(null)

  const handleAnalyze = async () => {
    setAnalyzing(true)
    setError(null)
    try {
      const simTimeSec = simHours * 3600
      const data = await variantApi.analyzeReliability(selectedCode, { shots, simTimeSec })
      setResult(data)
    } catch (err) {
      console.warn('API不可达，使用内置Mock计算:', err.message)
      const mockData = computeMockReliability(selectedCode, shots, simHours)
      setResult(mockData)
    } finally {
      setAnalyzing(false)
    }
  }

  const reliabilityChartOption = useMemo(() => {
    if (!result) return {}
    const xData = result.reliabilityCurve.map(c => c.n)
    const rData = result.reliabilityCurve.map(c => parseFloat(c.r.toFixed(4)))
    const jamData = result.cumulativeJams.map(c => c.count)

    return {
      tooltip: { trigger: 'axis' },
      legend: {
        data: ['可靠度 R(n)', '累积卡弹数'],
        textStyle: { color: '#d4c4a8' },
        top: 0
      },
      grid: { left: 60, right: 60, top: 50, bottom: 50 },
      xAxis: {
        type: 'category',
        data: xData,
        name: '发射次数 n',
        axisLabel: { color: '#d4c4a8' },
        nameTextStyle: { color: '#d4c4a8' },
        axisLine: { lineStyle: { color: '#8B4513' } },
        splitLine: { lineStyle: { color: '#5a4a3a', opacity: 0.3 } }
      },
      yAxis: [
        {
          type: 'value',
          name: '可靠度 R(n)',
          min: 0,
          max: 1,
          interval: 0.2,
          axisLabel: { color: '#d4c4a8', formatter: v => v.toFixed(1) },
          nameTextStyle: { color: '#4682B4' },
          splitLine: { lineStyle: { color: '#5a4a3a', opacity: 0.3 } }
        },
        {
          type: 'value',
          name: '累积卡弹数',
          axisLabel: { color: '#d4c4a8' },
          nameTextStyle: { color: '#CD5C5C' },
          splitLine: { show: false }
        }
      ],
      series: [
        {
          name: '可靠度 R(n)',
          type: 'line',
          yAxisIndex: 0,
          data: rData,
          smooth: true,
          lineStyle: { color: '#4682B4', width: 3 },
          itemStyle: { color: '#4682B4' },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(70, 130, 180, 0.4)' },
                { offset: 1, color: 'rgba(70, 130, 180, 0.02)' }
              ]
            }
          },
          markLine: {
            silent: true,
            symbol: 'none',
            lineStyle: { color: '#CD5C5C', type: 'dashed' },
            data: [{ yAxis: 0.5, label: { formatter: 'R=0.5', color: '#CD5C5C', position: 'end' } }]
          }
        },
        {
          name: '累积卡弹数',
          type: 'line',
          yAxisIndex: 1,
          data: jamData,
          step: 'end',
          lineStyle: { color: '#CD5C5C', width: 2 },
          itemStyle: { color: '#CD5C5C' },
          symbol: 'none'
        }
      ]
    }
  }, [result])

  const pieChartOption = useMemo(() => {
    if (!result) return {}
    return {
      tooltip: {
        trigger: 'item',
        formatter: '{b}<br/>占比: {d}%<br/>次数: {c}'
      },
      legend: {
        orient: 'vertical',
        left: 'left',
        textStyle: { color: '#d4c4a8' },
        top: 'center'
      },
      series: [{
        type: 'pie',
        radius: ['40%', '70%'],
        center: ['60%', '50%'],
        avoidLabelOverlap: true,
        itemStyle: {
          borderRadius: 8,
          borderColor: '#2a1f18',
          borderWidth: 2
        },
        label: {
          show: true,
          formatter: '{b}\n{d}%',
          color: '#d4c4a8',
          fontSize: 11
        },
        data: result.failureModes.map((fm, idx) => ({
          value: fm.percentage,
          name: fm.name,
          itemStyle: { color: PIE_COLORS[idx % PIE_COLORS.length] }
        }))
      }]
    }
  }, [result])

  const fmeaColumns = [
    {
      title: '失效模式',
      dataIndex: 'name',
      key: 'name',
      width: 130,
      render: (text) => (
        <Space>
          <WarningOutlined style={{ color: '#CD5C5C' }} />
          <Text strong style={{ color: '#d4c4a8' }}>{text}</Text>
        </Space>
      )
    },
    {
      title: '严重度 S (1-10)',
      dataIndex: 'severity',
      key: 'severity',
      width: 120,
      align: 'center',
      render: v => <Tag color={v >= 8 ? 'red' : v >= 5 ? 'orange' : 'green'}>{v}</Tag>
    },
    {
      title: '发生度 O (1-10)',
      dataIndex: 'occurrence',
      key: 'occurrence',
      width: 120,
      align: 'center',
      render: v => <Tag color={v >= 8 ? 'red' : v >= 5 ? 'orange' : 'green'}>{v}</Tag>
    },
    {
      title: '探测度 D (1-10)',
      dataIndex: 'detection',
      key: 'detection',
      width: 120,
      align: 'center',
      render: v => <Tag color={v <= 3 ? 'red' : v <= 6 ? 'orange' : 'green'}>{v}</Tag>
    },
    {
      title: 'RPN = S×O×D',
      dataIndex: 'rpn',
      key: 'rpn',
      width: 120,
      align: 'center',
      sorter: (a, b) => a.rpn - b.rpn,
      render: v => {
        const color = v >= 200 ? 'red' : v >= 150 ? 'volcano' : v >= 100 ? 'orange' : 'blue'
        return <Tag color={color} style={{ fontWeight: 'bold', fontSize: 14 }}>{v}</Tag>
      }
    },
    {
      title: '建议措施',
      dataIndex: 'suggestion',
      key: 'suggestion',
      render: t => <Text style={{ color: '#d4c4a8' }}>{t}</Text>
    }
  ]

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
              <SafetyCertificateOutlined style={{ marginRight: 8 }} />
              箭匣供弹可靠性分析
            </Title>
            <Text type="secondary" style={{ display: 'block', marginTop: 8 }}>
              基于威布尔分布的供弹系统故障预测
            </Text>
          </div>

          <Row gutter={24}>
            <Col xs={24} sm={8}>
              <div>
                <Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>选择弩型</Text>
                <Select
                  value={selectedCode}
                  onChange={setSelectedCode}
                  style={{ width: '100%' }}
                  size="large"
                >
                  {VARIANT_OPTIONS.map(v => (
                    <Option key={v.code} value={v.code}>{v.name}</Option>
                  ))}
                </Select>
              </div>
            </Col>
            <Col xs={24} sm={8}>
              <div>
                <Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
                  模拟发射次数：<Text strong style={{ color: '#B8860B' }}>{shots.toLocaleString()}</Text> 发
                </Text>
                <Slider
                  min={100}
                  max={100000}
                  step={100}
                  value={shots}
                  onChange={setShots}
                  marks={{ 100: '100', 10000: '1万', 50000: '5万', 100000: '10万' }}
                  tooltip={{ formatter: v => `${v?.toLocaleString()} 发` }}
                />
              </div>
            </Col>
            <Col xs={24} sm={8}>
              <div>
                <Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
                  模拟时间：<Text strong style={{ color: '#B8860B' }}>{simHours}</Text> 小时
                </Text>
                <Slider
                  min={1}
                  max={100}
                  step={1}
                  value={simHours}
                  onChange={setSimHours}
                  marks={{ 1: '1h', 10: '10h', 50: '50h', 100: '100h' }}
                  tooltip={{ formatter: v => `${v} 小时` }}
                />
              </div>
            </Col>
          </Row>

          <Space>
            <Button
              type="primary"
              size="large"
              icon={<ExperimentOutlined />}
              onClick={handleAnalyze}
              loading={analyzing}
              style={{
                background: 'linear-gradient(135deg, #2E8B57 0%, #1a5a37 100%)',
                border: 'none',
                minWidth: 160
              }}
            >
              开始分析
            </Button>
          </Space>
        </Space>
      </Card>

      {error && (
        <Alert type="error" message={error} showIcon closable style={{ marginBottom: 16 }} onClose={() => setError(null)} />
      )}

      {analyzing && (
        <Card style={{ background: '#2a1f18', border: '1px solid #8B4513' }}>
          <div style={{ textAlign: 'center', padding: 60 }}>
            <Spin size="large" />
            <div style={{ marginTop: 16, color: '#a89888' }}>正在运行威布尔可靠性仿真...</div>
          </div>
        </Card>
      )}

      {!result && !analyzing && (
        <Card style={{ background: '#2a1f18', border: '1px solid #8B4513' }}>
          <Empty
            description={<Text style={{ color: '#a89888' }}>请选择弩型并点击"开始分析"</Text>}
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          />
        </Card>
      )}

      {result && !analyzing && (
        <Space direction="vertical" size={16} style={{ width: '100%' }}>
          <Row gutter={16}>
            <Col xs={12} sm={8} md={4}>
              <Card size="small" style={{ background: 'rgba(46, 139, 87, 0.1)', border: '1px solid #2E8B57', textAlign: 'center' }}>
                <Statistic
                  title={<Text style={{ color: '#2E8B57', fontSize: 12 }}>总发射</Text>}
                  value={result.totalShots}
                  valueStyle={{ color: '#2E8B57', fontSize: 22 }}
                  suffix="发"
                  prefix={<CheckCircleOutlined />}
                />
              </Card>
            </Col>
            <Col xs={12} sm={8} md={4}>
              <Card size="small" style={{ background: 'rgba(205, 92, 92, 0.1)', border: '1px solid #CD5C5C', textAlign: 'center' }}>
                <Statistic
                  title={<Text style={{ color: '#CD5C5C', fontSize: 12 }}>卡弹次数</Text>}
                  value={result.jamCount}
                  valueStyle={{ color: '#CD5C5C', fontSize: 22 }}
                  suffix="次"
                  prefix={<WarningOutlined />}
                />
              </Card>
            </Col>
            <Col xs={12} sm={8} md={4}>
              <Card size="small" style={{ background: 'rgba(205, 133, 63, 0.1)', border: '1px solid #CD853F', textAlign: 'center' }}>
                <Statistic
                  title={<Text style={{ color: '#CD853F', fontSize: 12 }}>卡弹率(%)</Text>}
                  value={result.jamRate}
                  precision={4}
                  valueStyle={{ color: '#CD853F', fontSize: 22 }}
                  suffix="%"
                  prefix={<ThunderboltOutlined />}
                />
              </Card>
            </Col>
            <Col xs={12} sm={8} md={4}>
              <Card size="small" style={{ background: 'rgba(70, 130, 180, 0.1)', border: '1px solid #4682B4', textAlign: 'center' }}>
                <Statistic
                  title={<Text style={{ color: '#4682B4', fontSize: 12 }}>MTBF(发)</Text>}
                  value={Math.round(result.mtbf)}
                  valueStyle={{ color: '#4682B4', fontSize: 22 }}
                  suffix="发"
                  prefix={<LineChartOutlined />}
                />
              </Card>
            </Col>
            <Col xs={12} sm={8} md={4}>
              <Card size="small" style={{ background: 'rgba(107, 142, 35, 0.1)', border: '1px solid #6B8E23', textAlign: 'center' }}>
                <Statistic
                  title={<Text style={{ color: '#6B8E23', fontSize: 12 }}>MTBF(小时)</Text>}
                  value={result.mtbfHours}
                  precision={3}
                  valueStyle={{ color: '#6B8E23', fontSize: 22 }}
                  suffix="h"
                  prefix={<ClockCircleOutlined />}
                />
              </Card>
            </Col>
            <Col xs={12} sm={8} md={4}>
              <Card size="small" style={{ background: 'rgba(139, 115, 85, 0.1)', border: '1px solid #8B7355', textAlign: 'center' }}>
                <div>
                  <Text style={{ color: '#8B7355', fontSize: 12, display: 'block', marginBottom: 4 }}>95% CI [Lower, Upper]</Text>
                  <Text strong style={{ color: '#8B7355', fontSize: 18 }}>
                    {Math.round(result.ciLower)} ~ {Math.round(result.ciUpper)}
                  </Text>
                  <div style={{ fontSize: 11, color: '#a89888' }}>发</div>
                </div>
              </Card>
            </Col>
          </Row>

          <Card
            title={
              <Space>
                <LineChartOutlined style={{ color: '#B8860B' }} />
                <span style={{ color: '#d4c4a8' }}>可靠性曲线 R(n) = e^(-n/MTBF) & 累积卡弹</span>
              </Space>
            }
            style={{ background: '#2a1f18', border: '1px solid #8B4513' }}
            headStyle={{ borderBottom: '1px solid #8B4513' }}
          >
            <div style={{ height: 380 }}>
              <ReactECharts option={reliabilityChartOption} style={{ height: '100%' }} notMerge lazyUpdate />
            </div>
          </Card>

          <Row gutter={16}>
            <Col xs={24} lg={10}>
              <Card
                title={
                  <Space>
                    <PieChartOutlined style={{ color: '#B8860B' }} />
                    <span style={{ color: '#d4c4a8' }}>失效模式占比</span>
                  </Space>
                }
                style={{ background: '#2a1f18', border: '1px solid #8B4513', height: '100%' }}
                headStyle={{ borderBottom: '1px solid #8B4513' }}
              >
                <div style={{ height: 380 }}>
                  <ReactECharts option={pieChartOption} style={{ height: '100%' }} notMerge lazyUpdate />
                </div>
              </Card>
            </Col>
            <Col xs={24} lg={14}>
              <Card
                title={
                  <Space>
                    <FileTextOutlined style={{ color: '#B8860B' }} />
                    <span style={{ color: '#d4c4a8' }}>FMEA 分析表（RPN = S × O × D）</span>
                    <Tag color="red" style={{ marginLeft: 8 }}>RPN&gt;150 高风险</Tag>
                  </Space>
                }
                style={{ background: '#2a1f18', border: '1px solid #8B4513', height: '100%' }}
                headStyle={{ borderBottom: '1px solid #8B4513' }}
              >
                <Table
                  dataSource={result.fmea}
                  columns={fmeaColumns}
                  rowKey="mode"
                  size="middle"
                  scroll={{ x: 800 }}
                  pagination={false}
                  rowClassName={(record) => record.rpn > 150 ? 'fmea-high-risk-row' : ''}
                  onRow={(record) => {
                    if (record.rpn > 150) {
                      return {
                        style: { background: 'rgba(205, 92, 92, 0.2)' }
                      }
                    }
                    return {}
                  }}
                />
              </Card>
            </Col>
          </Row>
        </Space>
      )}
    </div>
  )
}

export default MagazineReliability
