package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strconv"
    "strings"
    "sync"
)

type WorkflowNode struct {
    ID     string  `json:"id"`
    Name   string  `json:"name"`
    Status string  `json:"status"`
    Desc   string  `json:"desc"`
}

// ---- 现实风格工作流模板 ----
type wfTemplate struct {
    ID    string
    Name  string
    Desc  string
    Nodes map[string]string   // id -> name
    Layers [][]string         // ordered layers of node IDs
    Cond map[[2]string]string // optional conditional edges: (from,to) -> label
}

func wfFromTemplate(t wfTemplate) WorkflowResponse {
    // Build nodes
    nodes := make([]WorkflowNode, 0, len(t.Nodes))
    for id, name := range t.Nodes {
        nodes = append(nodes, WorkflowNode{ID: id, Name: name, Status: "pending", Desc: t.Desc})
    }
    // deterministic order: ensure stable by ID
    // simple bubble sort due to small sizes
    for i := 0; i < len(nodes)-1; i++ {
        for j := i + 1; j < len(nodes); j++ {
            if nodes[i].ID > nodes[j].ID { nodes[i], nodes[j] = nodes[j], nodes[i] }
        }
    }
    // statuses for sources
    if len(t.Layers) > 0 && len(t.Layers[0]) > 0 {
        setStatus := func(id, st string) {
            for i := range nodes { if nodes[i].ID == id { nodes[i].Status = st; return } }
        }
        setStatus(t.Layers[0][0], "success")
        if len(t.Layers[0]) > 1 { setStatus(t.Layers[0][1], "running") }
    }
    // Build edges: connect each node in layer L to 1-2 nodes in L+1 (round-robin)
    var edges []WorkflowEdge
    for li := 0; li < len(t.Layers)-1; li++ {
        from := t.Layers[li]
        to := t.Layers[li+1]
        if len(from) == 0 || len(to) == 0 { continue }
        for i, u := range from {
            v1 := to[i%len(to)]
            edges = append(edges, WorkflowEdge{From: u, To: v1})
            if len(to) > 1 {
                v2 := to[(i+1)%len(to)]
                if v2 != v1 {
                    edges = append(edges, WorkflowEdge{From: u, To: v2})
                }
            }
        }
    }
    // Conditional edges
    if t.Cond != nil {
        for k, label := range t.Cond {
            edges = append(edges, WorkflowEdge{From: k[0], To: k[1], Type: "conditional", Label: label})
        }
    }
    // Enforce max 2 sources by adding an edge from first layer to overflow if necessary
    indeg := map[string]int{}
    for _, n := range nodes { indeg[n.ID] = 0 }
    for _, e := range edges { indeg[e.To]++ }
    // collect zero indegree
    zero := []string{}
    for _, n := range nodes { if indeg[n.ID] == 0 { zero = append(zero, n.ID) } }
    if len(zero) > 2 && len(t.Layers) > 1 && len(t.Layers[0]) > 0 {
        anchor := t.Layers[0][0]
        for i := 2; i < len(zero); i++ {
            edges = append(edges, WorkflowEdge{From: anchor, To: zero[i]})
        }
    }
    return WorkflowResponse{Nodes: nodes, Edges: edges}
}

func realisticCatalog() []wfTemplate {
    // 20 realistic templates across common domains
    return []wfTemplate{
        {ID: "wf-4", Name: "数据接入与落地", Desc: "数据采集/落地/分区/清单",
            Nodes: map[string]string{
                "SRC": "API采集", "SQOOP": "Sqoop导入", "RAW": "原始层(HDFS)",
                "PART": "分区整理", "MANIFEST": "清单生成", "QC": "质量校验",
                "DW": "明细层装载", "ACK": "回执", "ERR": "异常告警",
            },
            Layers: [][]string{{"SRC", "SQOOP"}, {"RAW", "PART"}, {"MANIFEST", "QC"}, {"DW", "ACK", "ERR"}},
            Cond: map[[2]string]string{{"QC","ERR"}: "fail", {"QC","DW"}: "pass"},
        },
        {ID: "wf-5", Name: "离线ETL汇总", Desc: "ODS->DWD->DWS->ADS",
            Nodes: map[string]string{"ODS":"ODS装载","DWD":"明细加工","DIM":"维表构建","DWS":"汇总层","ADS":"应用层","CHK":"校验","REP":"报表导出"},
            Layers: [][]string{{"ODS"},{"DWD","DIM"},{"DWS"},{"ADS","CHK"},{"REP"}},
            Cond: map[[2]string]string{{"CHK","REP"}: "ok"},
        },
        {ID: "wf-6", Name: "模型训练流水线-二分类", Desc: "特征->训练->评估->注册",
            Nodes: map[string]string{"ING":"样本准备","FE":"特征工程","SPLIT":"训练/验证划分","TRAIN":"训练(GBDT)","EVAL":"评估","REG":"模型注册","EXPL":"可解释性","PUSH":"推送线上"},
            Layers: [][]string{{"ING"},{"FE"},{"SPLIT"},{"TRAIN"},{"EVAL","EXPL"},{"REG"},{"PUSH"}},
            Cond: map[[2]string]string{{"EVAL","REG"}: ">=0.8"},
        },
        {ID: "wf-7", Name: "CI/CD 构建发布", Desc: "构建->测试->发布",
            Nodes: map[string]string{"SCM":"拉取代码","BUILD":"构建镜像","UT":"单测","IT":"集成测试","SEC":"安全扫描","STG":"灰度","PRD":"生产发布","ROLL":"回滚"},
            Layers: [][]string{{"SCM"},{"BUILD"},{"UT","IT","SEC"},{"STG"},{"PRD","ROLL"}},
            Cond: map[[2]string]string{{"STG","PRD"}: "ok", {"STG","ROLL"}: "fail"},
        },
        {ID: "wf-8", Name: "实时采集与聚合", Desc: "Kafka->Flink->OLAP",
            Nodes: map[string]string{"ING":"Kafka采集","FLK":"Flink清洗","AGG":"实时聚合","OLAP":"OLAP写入","ALM":"告警","DLQ":"死信队列"},
            Layers: [][]string{{"ING"},{"FLK"},{"AGG"},{"OLAP","ALM","DLQ"}},
            Cond: map[[2]string]string{{"FLK","DLQ"}: "invalid"},
        },
        {ID: "wf-9", Name: "日志分析链路", Desc: "采集->解析->索引->可视化",
            Nodes: map[string]string{"COL":"Filebeat","PAR":"解析","IDX":"索引(ES)","DASH":"仪表盘","ALM":"告警"},
            Layers: [][]string{{"COL"},{"PAR"},{"IDX"},{"DASH","ALM"}},
            Cond: nil,
        },
        {ID: "wf-10", Name: "推荐-离线召回池", Desc: "行为->Embedding->ANN",
            Nodes: map[string]string{"UV":"行为归档","FE":"Embedding训练","ANN":"近邻索引","EXP":"探索实验","OUT":"召回池导出"},
            Layers: [][]string{{"UV"},{"FE"},{"ANN","EXP"},{"OUT"}},
            Cond: nil,
        },
        {ID: "wf-11", Name: "风控-特征与规则", Desc: "风控特征与规则生成",
            Nodes: map[string]string{"RAW":"交易拉链","FE":"特征抽取","SEL":"特征筛选","RULE":"规则编译","PUB":"规则发布","MON":"监控"},
            Layers: [][]string{{"RAW"},{"FE"},{"SEL","RULE"},{"PUB","MON"}},
            Cond: nil,
        },
        {ID: "wf-12", Name: "IoT 遥测处理", Desc: "时序入库与告警",
            Nodes: map[string]string{"GW":"网关接入","DEC":"解码","TS":"时序写入","QC":"质量检测","ALM":"阈值告警","REP":"日报"},
            Layers: [][]string{{"GW"},{"DEC"},{"TS","QC"},{"ALM","REP"}},
            Cond: map[[2]string]string{{"QC","ALM"}: "fail"},
        },
        {ID: "wf-13", Name: "订单履约", Desc: "下单->拣配->配送",
            Nodes: map[string]string{"CRT":"下单","PAY":"支付","INV":"库存锁定","PCK":"拣货","SHP":"发货","RCV":"签收","RET":"退货"},
            Layers: [][]string{{"CRT","PAY"},{"INV"},{"PCK"},{"SHP"},{"RCV","RET"}},
            Cond: map[[2]string]string{{"PAY","INV"}: "paid"},
        },
        {ID: "wf-14", Name: "支付清结算", Desc: "清分对账流水",
            Nodes: map[string]string{"COL":"收单","CHK":"对账","CLR":"清分","SET":"结算","ALM":"异常"},
            Layers: [][]string{{"COL"},{"CHK"},{"CLR"},{"SET","ALM"}},
            Cond: map[[2]string]string{{"CHK","ALM"}: "mismatch"},
        },
        {ID: "wf-15", Name: "供应链补货", Desc: "预测->补货->到货",
            Nodes: map[string]string{"FCST":"销量预测","PLAN":"补货计划","PO":"采购单","ASN":"到货通知","IN":"入库","ALM":"缺货告警"},
            Layers: [][]string{{"FCST"},{"PLAN"},{"PO"},{"ASN"},{"IN","ALM"}},
            Cond: nil,
        },
        {ID: "wf-16", Name: "A/B 实验评估", Desc: "实验拆分与评估",
            Nodes: map[string]string{"SPL":"流量拆分","COL":"指标采集","EVA":"统计检验","DEC":"决策","ROL":"回滚","ROLF":"跟进"},
            Layers: [][]string{{"SPL"},{"COL"},{"EVA"},{"DEC"},{"ROL","ROLF"}},
            Cond: map[[2]string]string{{"DEC","ROL"}: "bad", {"DEC","ROLF"}: "good"},
        },
        {ID: "wf-17", Name: "特征库构建", Desc: "批/流统一特征",
            Nodes: map[string]string{"RAW":"埋点归档","DIM":"维度填充","AGG":"窗口聚合","JOIN":"多源拼接","VAL":"校验","PUB":"发布"},
            Layers: [][]string{{"RAW"},{"DIM"},{"AGG","JOIN"},{"VAL"},{"PUB"}},
            Cond: nil,
        },
        {ID: "wf-18", Name: "模型上线与灰度", Desc: "模型打包/灰度/上线",
            Nodes: map[string]string{"PKG":"模型打包","IMG":"镜像构建","DEP":"部署","AB":"灰度","MON":"监控","ALR":"告警"},
            Layers: [][]string{{"PKG"},{"IMG"},{"DEP"},{"AB"},{"MON","ALR"}},
            Cond: map[[2]string]string{{"AB","MON"}: "ok", {"AB","ALR"}: "fail"},
        },
        {ID: "wf-19", Name: "监控与告警", Desc: "采集/规则/通知",
            Nodes: map[string]string{"SCR":"采集","TS":"聚合","RUL":"规则","NTF":"通知","TKT":"工单"},
            Layers: [][]string{{"SCR"},{"TS"},{"RUL"},{"NTF","TKT"}},
            Cond: nil,
        },
        {ID: "wf-20", Name: "数据质量检查", Desc: "完整性/唯一性/范围",
            Nodes: map[string]string{"IMP":"导入","SCM":"Schema检查","UNI":"唯一性","RNG":"范围","REP":"报告","BLK":"阻断"},
            Layers: [][]string{{"IMP"},{"SCM"},{"UNI","RNG"},{"REP","BLK"}},
            Cond: map[[2]string]string{{"SCM","BLK"}: "invalid"},
        },
        {ID: "wf-21", Name: "客户开户流程", Desc: "审查/签约/开户",
            Nodes: map[string]string{"APPLY":"提交申请","KYC":"身份审查","RISK":"风险评估","SIGN":"签约","OPEN":"开户","REJ":"拒绝"},
            Layers: [][]string{{"APPLY"},{"KYC","RISK"},{"SIGN","REJ"},{"OPEN"}},
            Cond: map[[2]string]string{{"KYC","REJ"}: "fail", {"RISK","REJ"}: "high"},
        },
        {ID: "wf-22", Name: "账单出具", Desc: "对账/计费/出单",
            Nodes: map[string]string{"COL":"采集用量","RAT":"计费","INV":"出账","NOT":"通知","DIS":"争议"},
            Layers: [][]string{{"COL"},{"RAT"},{"INV"},{"NOT","DIS"}},
            Cond: nil,
        },
        {ID: "wf-23", Name: "流失预测", Desc: "训练->评估->投产",
            Nodes: map[string]string{"ETL":"数据准备","FE":"特征","TR":"训练","EV":"评估","REG":"注册","EXP":"实验平台"},
            Layers: [][]string{{"ETL"},{"FE"},{"TR"},{"EV"},{"REG","EXP"}},
            Cond: map[[2]string]string{{"EV","REG"}: ">=0.75"},
        },
    }
}

// ---- 生成型工作流（复杂，5~30 节点） ----
// 简易可重复伪随机
type prng struct{ state uint64 }

func (p *prng) next() uint64 { // xorshift64*
    x := p.state
    x ^= x >> 12
    x ^= x << 25
    x ^= x >> 27
    p.state = x
    return x * 2685821657736338717
}
func (p *prng) nextInt(n int) int {
    if n <= 0 { return 0 }
    return int((p.next() >> 32) % uint64(n))
}

func hashString(s string) uint64 {
    var h uint64 = 1469598103934665603 // FNV offset
    for i := 0; i < len(s); i++ {
        h ^= uint64(s[i])
        h *= 1099511628211
    }
    if h == 0 { h = 1 }
    return h
}

func genWorkflow(id string) WorkflowResponse {
    rng := &prng{state: hashString(id)}
    // 节点数量：5~30
    nodeCount := 5 + rng.nextInt(26)
    // 层数：3~6
    layerCount := 3 + rng.nextInt(4)
    layers := make([][]int, layerCount)
    // 先分配每层至少一个节点
    for i := 0; i < layerCount; i++ {
        layers[i] = []int{}
    }
    for i := 0; i < nodeCount; i++ {
        li := rng.nextInt(layerCount)
        layers[li] = append(layers[li], i)
    }
    // 确保首层和末层非空
    if len(layers[0]) == 0 {
        layers[0] = append(layers[0], 0)
    }
    if len(layers[layerCount-1]) == 0 {
        layers[layerCount-1] = append(layers[layerCount-1], nodeCount-1)
    }
    // 约束：无入边节点（源头）<= 2
    if len(layers[0]) > 2 {
        // 将多余的首层节点移动到后续层（优先第二层）
        overflow := layers[0][2:]
        layers[0] = layers[0][:2]
        targetLayer := 1
        if targetLayer >= layerCount { targetLayer = layerCount - 1 }
        for _, idx := range overflow {
            // 避免与末层冲突，若只有两层则仍放在末层
            li := targetLayer
            if layerCount > 2 {
                li = 1 + rng.nextInt(layerCount-1) // 分布在 1..layerCount-1
            }
            layers[li] = append(layers[li], idx)
        }
    }
    // 生成节点
    statuses := []string{"pending", "running", "success", "failed"}
    nodes := make([]WorkflowNode, nodeCount)
    for i := 0; i < nodeCount; i++ {
        nodes[i] = WorkflowNode{
            ID:     fmt.Sprintf("N%d", i+1),
            Name:   fmt.Sprintf("节点%02d", i+1),
            Status: statuses[rng.nextInt(len(statuses))],
            Desc:   "自动生成",
        }
    }
    // 确保至少有一个起点 success（若有两个起点，第二个设为 running 提示并行）
    if len(layers[0]) > 0 {
        nodes[layers[0][0]].Status = "success"
        if len(layers[0]) > 1 && nodes[layers[0][1]].Status == "pending" {
            nodes[layers[0][1]].Status = "running"
        }
    }
    // 生成边：只在相邻层连边，避免环
    var edges []WorkflowEdge
    for li := 0; li < layerCount-1; li++ {
        fromLayer := layers[li]
        toLayer := layers[li+1]
        if len(fromLayer) == 0 || len(toLayer) == 0 { continue }
        for _, u := range fromLayer {
            // 每个节点 1~3 条出边
            outDeg := 1 + rng.nextInt(3)
            used := map[int]bool{}
            for k := 0; k < outDeg; k++ {
                vIdx := rng.nextInt(len(toLayer))
                if used[vIdx] { continue }
                used[vIdx] = true
                v := toLayer[vIdx]
                e := WorkflowEdge{From: nodes[u].ID, To: nodes[v].ID}
                // 部分条件边
                if rng.nextInt(4) == 0 {
                    e.Type = "conditional"
                    e.Label = []string{"if ok", "rule", "split", "merge"}[rng.nextInt(4)]
                }
                edges = append(edges, e)
            }
        }
    }
    // 保证末层可达（如仍无入边，给它补一条）
    for _, v := range layers[layerCount-1] {
        hasIn := false
        for _, e := range edges {
            if e.To == nodes[v].ID { hasIn = true; break }
        }
        if !hasIn && layerCount >= 2 && len(layers[layerCount-2]) > 0 {
            u := layers[layerCount-2][rng.nextInt(len(layers[layerCount-2]))]
            edges = append(edges, WorkflowEdge{From: nodes[u].ID, To: nodes[v].ID})
        }
    }
    // 全局约束：无入边节点总数 <= 2
    // 计算入度
    indeg := make(map[string]int, len(nodes))
    for i := range nodes { indeg[nodes[i].ID] = 0 }
    for _, e := range edges { indeg[e.To]++ }
    // 收集零入度节点索引
    type zi struct{ idx int; layer int }
    var zeros []zi
    // 快速查询每个节点所在层
    nodeLayer := make(map[int]int, len(nodes))
    for li := range layers {
        for _, idx := range layers[li] { nodeLayer[idx] = li }
    }
    for i := range nodes {
        if indeg[nodes[i].ID] == 0 { zeros = append(zeros, zi{idx: i, layer: nodeLayer[i]}) }
    }
    if len(zeros) > 2 {
        // 按层升序排序，尽量保留更靠前层作为源头
        for i := 0; i < len(zeros)-1; i++ {
            for j := i + 1; j < len(zeros); j++ {
                if zeros[i].layer > zeros[j].layer {
                    zeros[i], zeros[j] = zeros[j], zeros[i]
                }
            }
        }
        // 保留前两个作为源头，其余补一条来自前一层的边
        for k := 2; k < len(zeros); k++ {
            z := zeros[k]
            li := z.layer
            if li <= 0 { continue }
            prev := layers[li-1]
            if len(prev) == 0 { continue }
            u := prev[rng.nextInt(len(prev))]
            e := WorkflowEdge{From: nodes[u].ID, To: nodes[z.idx].ID}
            edges = append(edges, e)
            indeg[nodes[z.idx].ID]++
        }
    }
    return WorkflowResponse{Nodes: nodes, Edges: edges}
}

// Different mock DAGs by workflow id
func mockWorkflowByID(id string) WorkflowResponse {
    switch id {
    case "wf-1":
        // 训练流水线（较复杂，接近原有示例）
        return mockWorkflow()
    case "wf-2":
        // 数据质量检查（简化菱形结构）
        nodes := []WorkflowNode{
            {ID: "A", Name: "拉取批次", Status: "success", Desc: "拉取待校验批次"},
            {ID: "B", Name: "Schema校验", Status: "running", Desc: "字段与类型"},
            {ID: "C", Name: "唯一性检测", Status: "pending", Desc: "主键重复"},
            {ID: "D", Name: "空值检测", Status: "pending", Desc: "必填空值"},
            {ID: "E", Name: "质量评分", Status: "pending", Desc: "综合评分"},
            {ID: "F", Name: "生成报告", Status: "pending", Desc: "输出报表"},
        }
        edges := []WorkflowEdge{
            {From: "A", To: "B"}, {From: "A", To: "C"}, {From: "A", To: "D"},
            {From: "B", To: "E"}, {From: "C", To: "E"}, {From: "D", To: "E"},
            {From: "E", To: "F"},
        }
        return WorkflowResponse{Nodes: nodes, Edges: edges}
    case "wf-3":
        // 特征抽取与选择（线性+分叉）
        nodes := []WorkflowNode{
            {ID: "A", Name: "加载原始数据", Status: "success", Desc: "ODS"},
            {ID: "B", Name: "清洗", Status: "success", Desc: "缺失/异常"},
            {ID: "C", Name: "特征抽取", Status: "running", Desc: "统计/频次"},
            {ID: "D", Name: "特征筛选", Status: "pending", Desc: "过滤法"},
            {ID: "E", Name: "降维",   Status: "pending", Desc: "PCA"},
            {ID: "F", Name: "导出特征", Status: "pending", Desc: "保存到特征库"},
        }
        edges := []WorkflowEdge{
            {From: "A", To: "B"}, {From: "B", To: "C"},
            {From: "C", To: "D"}, {From: "C", To: "E"},
            {From: "D", To: "F"}, {From: "E", To: "F"},
        }
        return WorkflowResponse{Nodes: nodes, Edges: edges}
    case "wf-4", "wf-5", "wf-6", "wf-7", "wf-8", "wf-9", "wf-10", "wf-11", "wf-12", "wf-13",
        "wf-14", "wf-15", "wf-16", "wf-17", "wf-18", "wf-19", "wf-20", "wf-21", "wf-22", "wf-23":
        // 优先从现实模板返回
        for _, t := range realisticCatalog() {
            if t.ID == id { return wfFromTemplate(t) }
        }
        return genWorkflow(id)
    default:
        // 其他 id 按生成规则返回
        return genWorkflow(id)
    }
}

type WorkflowEdge struct {
    From  string `json:"from"`
    To    string `json:"to"`
    Type  string `json:"type,omitempty"`
    Label string `json:"label,omitempty"`
}

type WorkflowResponse struct {
    Nodes []WorkflowNode `json:"nodes"`
    Edges []WorkflowEdge `json:"edges"`
}

// ---- In-memory store for user-created workflows ----
var (
    createdMu        sync.RWMutex
    createdWorkflows = map[string]WorkflowResponse{}
    createdSummaries = map[string]WorkflowSummary{}
    createdSeq       = 0
)

type CreateWorkflowRequest struct {
    ID     string           `json:"id"`
    Name   string           `json:"name"`
    Desc   string           `json:"desc"`
    Nodes  []WorkflowNode   `json:"nodes"`
    Edges  []WorkflowEdge   `json:"edges"`
}

// Workflow summary for listing
type WorkflowSummary struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    Status string `json:"status"`
    Desc   string `json:"desc"`
}

// mock workflow data (same layout as the current Vue demo)
func mockWorkflow() WorkflowResponse {
    nodes := []WorkflowNode{
        {ID: "A", Name: "任务A", Status: "success", Desc: "数据准备"},
        {ID: "K", Name: "任务K", Status: "pending", Desc: "数据采样"},
        {ID: "B", Name: "任务B", Status: "running", Desc: "数据清洗"},
        {ID: "C", Name: "任务C", Status: "success", Desc: "特征工程"},
        {ID: "L", Name: "任务L", Status: "pending", Desc: "规则生成"},
        {ID: "O", Name: "任务O", Status: "pending", Desc: "数据质量校验"},
        {ID: "D", Name: "任务D", Status: "pending", Desc: "模型训练1"},
        {ID: "E", Name: "任务E", Status: "pending", Desc: "模型训练2"},
        {ID: "P", Name: "任务P", Status: "pending", Desc: "特征选择"},
        {ID: "F", Name: "任务F", Status: "failed", Desc: "模型训练3"},
        {ID: "G", Name: "任务G", Status: "pending", Desc: "模型训练4"},
        {ID: "H", Name: "任务H", Status: "success", Desc: "评估1"},
        {ID: "I", Name: "任务I", Status: "pending", Desc: "评估2"},
        {ID: "M", Name: "任务M", Status: "pending", Desc: "可视化报告"},
        {ID: "N", Name: "任务N", Status: "pending", Desc: "在线部署"},
        {ID: "J", Name: "任务J", Status: "pending", Desc: "结果汇总"},
        {ID: "Q", Name: "任务Q", Status: "pending", Desc: "超参搜索"},
        {ID: "R", Name: "任务R", Status: "pending", Desc: "A/B 测试"},
        {ID: "S", Name: "任务S", Status: "pending", Desc: "回滚策略"},
        {ID: "T", Name: "任务T", Status: "pending", Desc: "上线审核"},
        {ID: "U", Name: "任务U", Status: "pending", Desc: "告警与监控"},
    }
    edges := []WorkflowEdge{
        {From: "A", To: "K", Type: "conditional", Label: "if sample"},
        {From: "A", To: "B"},
        {From: "A", To: "C"},
        {From: "B", To: "D"},
        {From: "B", To: "E"},
        {From: "C", To: "F"},
        {From: "C", To: "G"},
        {From: "K", To: "D", Type: "conditional", Label: "small set"},
        {From: "L", To: "G", Type: "conditional", Label: "rule add"},
        {From: "D", To: "H"},
        {From: "E", To: "H"},
        {From: "F", To: "I"},
        {From: "G", To: "I"},
        {From: "H", To: "J"},
        {From: "I", To: "J"},
        {From: "H", To: "M", Type: "conditional", Label: "report"},
        {From: "I", To: "N", Type: "conditional", Label: "deploy"},
        {From: "A", To: "L", Type: "conditional", Label: "if rules"},
        {From: "B", To: "O", Type: "conditional", Label: "dq check"},
        {From: "O", To: "D"},
        {From: "C", To: "P", Type: "conditional", Label: "feature select"},
        {From: "P", To: "G"},
        {From: "P", To: "F", Type: "conditional", Label: "drop noisy"},
        {From: "J", To: "Q"},
        {From: "Q", To: "R"},
        {From: "Q", To: "S"},
        {From: "R", To: "T"},
        {From: "S", To: "T"},
        {From: "T", To: "U"},
        {From: "N", To: "T", Type: "conditional", Label: "pre-prod"},
    }
    return WorkflowResponse{Nodes: nodes, Edges: edges}
}

func workflowHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
        return
    }
    wf := mockWorkflow()
    writeJSON(w, http.StatusOK, wf)
}

// mock list of workflows
func mockWorkflowList() []WorkflowSummary {
    list := []WorkflowSummary{
        {ID: "wf-1", Name: "模型训练流水线", Status: "running", Desc: "每日训练任务"},
        {ID: "wf-2", Name: "数据质量检查", Status: "success", Desc: "入库校验"},
        {ID: "wf-3", Name: "特征抽取与选择", Status: "pending", Desc: "离线批处理"},
    }
    // 追加现实风格模板（wf-4 至 wf-23）
    statuses := []string{"pending", "running", "success"}
    rc := realisticCatalog()
    for i, t := range rc {
        s := statuses[((i+4)*7)%len(statuses)]
        list = append(list, WorkflowSummary{ID: t.ID, Name: t.Name, Status: s, Desc: t.Desc})
    }
    return list
}

// GET /api/v1/workflows (list), POST /api/v1/workflows (create)
func workflowsCollectionHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        getWorkflowsList(w, r)
    case http.MethodPost:
        createWorkflow(w, r)
    default:
        writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
    }
}

// GET /api/v1/workflows/{id}, PUT /api/v1/workflows/{id}, DELETE /api/v1/workflows/{id}
func workflowResourceHandler(w http.ResponseWriter, r *http.Request) {
    // 提取工作流 ID
    path := r.URL.Path
    prefix := "/api/v1/workflows/"
    if !strings.HasPrefix(path, prefix) {
        writeJSON(w, http.StatusNotFound, map[string]string{"error": "Not found"})
        return
    }
    id := strings.TrimPrefix(path, prefix)
    if id == "" || strings.Contains(id, "/") {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid workflow ID"})
        return
    }

    switch r.Method {
    case http.MethodGet:
        getWorkflow(w, r, id)
    case http.MethodPut:
        updateWorkflow(w, r, id)
    case http.MethodDelete:
        deleteWorkflow(w, r, id)
    default:
        writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
    }
}


func getWorkflowsList(w http.ResponseWriter, r *http.Request) {
    list := mockWorkflowList()
    // 合并用户创建的工作流摘要
    createdMu.RLock()
    for _, s := range createdSummaries {
        list = append(list, s)
    }
    createdMu.RUnlock()

    // 分页参数
    page := 1
    pageSize := 20
    q := r.URL.Query()
    if p := q.Get("page"); p != "" {
        if v, err := strconv.Atoi(p); err == nil && v > 0 {
            page = v
        }
    }
    if ps := q.Get("pageSize"); ps != "" {
        if v, err := strconv.Atoi(ps); err == nil && v > 0 {
            pageSize = v
        }
    }

    start := (page - 1) * pageSize
    end := start + pageSize
    if start > len(list) {
        start = len(list)
    }
    if end > len(list) {
        end = len(list)
    }
    paged := list[start:end]

    resp := map[string]interface{}{
        "workflows": paged,
        "total":     len(list),
        "page":      page,
        "pageSize":  pageSize,
    }
    writeJSON(w, http.StatusOK, resp)
}

func createWorkflow(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
        return
    }

    var req CreateWorkflowRequest
    if err := json.Unmarshal(body, &req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
        return
    }

    // 验证必填字段
    if len(req.Nodes) == 0 {
        writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "At least one node is required"})
        return
    }

    // 验证节点 ID 唯一性
    ids := map[string]bool{}
    for i, n := range req.Nodes {
        if strings.TrimSpace(n.ID) == "" {
            writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "Node ID is required"})
            return
        }
        if ids[n.ID] {
            writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "Duplicate node ID: " + n.ID})
            return
        }
        ids[n.ID] = true
        // 设置默认状态
        if n.Status == "" {
            req.Nodes[i].Status = "pending"
        }
    }

    // 过滤无效边
    edgeOK := make([]WorkflowEdge, 0, len(req.Edges))
    for _, e := range req.Edges {
        if e.From == "" || e.To == "" {
            continue
        }
        if !ids[e.From] || !ids[e.To] {
            continue
        }
        edgeOK = append(edgeOK, e)
    }

    createdMu.Lock()
    defer createdMu.Unlock()

    // 生成或验证 ID
    id := strings.TrimSpace(req.ID)
    if id == "" {
        createdSeq++
        id = fmt.Sprintf("user-wf-%d", createdSeq)
    }

    // 检查 ID 冲突
    if _, ok := createdWorkflows[id]; ok {
        writeJSON(w, http.StatusConflict, map[string]string{"error": "Workflow ID already exists"})
        return
    }

    // 创建工作流
    wf := WorkflowResponse{Nodes: req.Nodes, Edges: edgeOK}
    createdWorkflows[id] = wf

    name := strings.TrimSpace(req.Name)
    if name == "" {
        name = id
    }
    desc := strings.TrimSpace(req.Desc)
    status := "pending"
    if len(req.Nodes) > 0 {
        status = req.Nodes[0].Status
    }

    createdSummaries[id] = WorkflowSummary{
        ID:     id,
        Name:   name,
        Desc:   desc,
        Status: status,
    }

    // 返回创建的资源
    response := map[string]interface{}{
        "id":       id,
        "name":     name,
        "desc":     desc,
        "status":   status,
        "nodes":    len(req.Nodes),
        "edges":    len(edgeOK),
    }
    writeJSON(w, http.StatusCreated, response)
}

func getWorkflow(w http.ResponseWriter, r *http.Request, id string) {
    // 优先返回用户创建的工作流
    createdMu.RLock()
    if wf, ok := createdWorkflows[id]; ok {
        createdMu.RUnlock()
        writeJSON(w, http.StatusOK, wf)
        return
    }
    createdMu.RUnlock()

    // 回退到 mock 数据
    wf := mockWorkflowByID(id)
    writeJSON(w, http.StatusOK, wf)
}

func updateWorkflow(w http.ResponseWriter, r *http.Request, id string) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
        return
    }

    var req CreateWorkflowRequest
    if err := json.Unmarshal(body, &req); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
        return
    }

    createdMu.Lock()
    defer createdMu.Unlock()

    // 检查工作流是否存在（仅支持更新用户创建的工作流）
    if _, ok := createdWorkflows[id]; !ok {
        writeJSON(w, http.StatusNotFound, map[string]string{"error": "Workflow not found or not editable"})
        return
    }

    // 验证节点（与创建时相同的逻辑）
    if len(req.Nodes) == 0 {
        writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "At least one node is required"})
        return
    }

    ids := map[string]bool{}
    for i, n := range req.Nodes {
        if strings.TrimSpace(n.ID) == "" {
            writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "Node ID is required"})
            return
        }
        if ids[n.ID] {
            writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "Duplicate node ID: " + n.ID})
            return
        }
        ids[n.ID] = true
        if n.Status == "" {
            req.Nodes[i].Status = "pending"
        }
    }

    // 过滤无效边
    edgeOK := make([]WorkflowEdge, 0, len(req.Edges))
    for _, e := range req.Edges {
        if e.From == "" || e.To == "" {
            continue
        }
        if !ids[e.From] || !ids[e.To] {
            continue
        }
        edgeOK = append(edgeOK, e)
    }

    // 更新工作流
    wf := WorkflowResponse{Nodes: req.Nodes, Edges: edgeOK}
    createdWorkflows[id] = wf

    // 更新摘要
    name := strings.TrimSpace(req.Name)
    if name == "" {
        name = id
    }
    desc := strings.TrimSpace(req.Desc)
    status := "pending"
    if len(req.Nodes) > 0 {
        status = req.Nodes[0].Status
    }

    createdSummaries[id] = WorkflowSummary{
        ID:     id,
        Name:   name,
        Desc:   desc,
        Status: status,
    }

    writeJSON(w, http.StatusOK, wf)
}

func deleteWorkflow(w http.ResponseWriter, r *http.Request, id string) {
    createdMu.Lock()
    defer createdMu.Unlock()

    // 检查工作流是否存在（仅支持删除用户创建的工作流）
    if _, ok := createdWorkflows[id]; !ok {
        writeJSON(w, http.StatusNotFound, map[string]string{"error": "Workflow not found or not deletable"})
        return
    }

    // 删除工作流和摘要
    delete(createdWorkflows, id)
    delete(createdSummaries, id)

    writeJSON(w, http.StatusNoContent, nil)
}
