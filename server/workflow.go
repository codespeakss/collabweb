package main

import (
    "net/http"
    "strings"
)

type WorkflowNode struct {
    ID     string  `json:"id"`
    Name   string  `json:"name"`
    X      float64 `json:"x"`
    Y      float64 `json:"y"`
    Status string  `json:"status"`
    Desc   string  `json:"desc"`
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
            {ID: "A", Name: "拉取批次", X: 200, Y: 80, Status: "success", Desc: "拉取待校验批次"},
            {ID: "B", Name: "Schema校验", X: 80, Y: 200, Status: "running", Desc: "字段与类型"},
            {ID: "C", Name: "唯一性检测", X: 200, Y: 200, Status: "pending", Desc: "主键重复"},
            {ID: "D", Name: "空值检测", X: 320, Y: 200, Status: "pending", Desc: "必填空值"},
            {ID: "E", Name: "质量评分", X: 200, Y: 320, Status: "pending", Desc: "综合评分"},
            {ID: "F", Name: "生成报告", X: 200, Y: 440, Status: "pending", Desc: "输出报表"},
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
            {ID: "A", Name: "加载原始数据", X: 120, Y: 80, Status: "success", Desc: "ODS"},
            {ID: "B", Name: "清洗", X: 120, Y: 180, Status: "success", Desc: "缺失/异常"},
            {ID: "C", Name: "特征抽取", X: 120, Y: 280, Status: "running", Desc: "统计/频次"},
            {ID: "D", Name: "特征筛选", X: 40,  Y: 380, Status: "pending", Desc: "过滤法"},
            {ID: "E", Name: "降维",   X: 200, Y: 380, Status: "pending", Desc: "PCA"},
            {ID: "F", Name: "导出特征", X: 120, Y: 480, Status: "pending", Desc: "保存到特征库"},
        }
        edges := []WorkflowEdge{
            {From: "A", To: "B"}, {From: "B", To: "C"},
            {From: "C", To: "D"}, {From: "C", To: "E"},
            {From: "D", To: "F"}, {From: "E", To: "F"},
        }
        return WorkflowResponse{Nodes: nodes, Edges: edges}
    default:
        // 未知 id 使用一个极简示例
        nodes := []WorkflowNode{
            {ID: "A", Name: "开始", X: 120, Y: 80,  Status: "success", Desc: "起点"},
            {ID: "B", Name: "处理", X: 120, Y: 200, Status: "running", Desc: "任务"},
            {ID: "C", Name: "结束", X: 120, Y: 320, Status: "pending", Desc: "终点"},
        }
        edges := []WorkflowEdge{{From: "A", To: "B"}, {From: "B", To: "C"}}
        return WorkflowResponse{Nodes: nodes, Edges: edges}
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
        {ID: "A", Name: "任务A", X: 390, Y: 80, Status: "success", Desc: "数据准备"},
        {ID: "K", Name: "任务K", X: 110, Y: 180, Status: "pending", Desc: "数据采样"},
        {ID: "B", Name: "任务B", X: 250, Y: 180, Status: "running", Desc: "数据清洗"},
        {ID: "C", Name: "任务C", X: 530, Y: 180, Status: "success", Desc: "特征工程"},
        {ID: "L", Name: "任务L", X: 670, Y: 180, Status: "pending", Desc: "规则生成"},
        {ID: "O", Name: "任务O", X: 30,  Y: 280, Status: "pending", Desc: "数据质量校验"},
        {ID: "D", Name: "任务D", X: 170, Y: 280, Status: "pending", Desc: "模型训练1"},
        {ID: "E", Name: "任务E", X: 310, Y: 280, Status: "pending", Desc: "模型训练2"},
        {ID: "P", Name: "任务P", X: 750, Y: 280, Status: "pending", Desc: "特征选择"},
        {ID: "F", Name: "任务F", X: 470, Y: 380, Status: "failed", Desc: "模型训练3"},
        {ID: "G", Name: "任务G", X: 610, Y: 380, Status: "pending", Desc: "模型训练4"},
        {ID: "H", Name: "任务H", X: 250, Y: 480, Status: "success", Desc: "评估1"},
        {ID: "I", Name: "任务I", X: 530, Y: 480, Status: "pending", Desc: "评估2"},
        {ID: "M", Name: "任务M", X: 110, Y: 580, Status: "pending", Desc: "可视化报告"},
        {ID: "N", Name: "任务N", X: 670, Y: 580, Status: "pending", Desc: "在线部署"},
        {ID: "J", Name: "任务J", X: 390, Y: 580, Status: "pending", Desc: "结果汇总"},
        {ID: "Q", Name: "任务Q", X: 390, Y: 680, Status: "pending", Desc: "超参搜索"},
        {ID: "R", Name: "任务R", X: 250, Y: 780, Status: "pending", Desc: "A/B 测试"},
        {ID: "S", Name: "任务S", X: 530, Y: 780, Status: "pending", Desc: "回滚策略"},
        {ID: "T", Name: "任务T", X: 390, Y: 880, Status: "pending", Desc: "上线审核"},
        {ID: "U", Name: "任务U", X: 390, Y: 960, Status: "pending", Desc: "告警与监控"},
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
    return []WorkflowSummary{
        {ID: "wf-1", Name: "模型训练流水线", Status: "running", Desc: "每日训练任务"},
        {ID: "wf-2", Name: "数据质量检查", Status: "success", Desc: "入库校验"},
        {ID: "wf-3", Name: "特征抽取与选择", Status: "pending", Desc: "离线批处理"},
    }
}

// GET /api/workflows -> list
func workflowsListHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
        return
    }
    list := mockWorkflowList()
    writeJSON(w, http.StatusOK, list)
}

// GET /api/workflows/{id} -> detail
func workflowDetailHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
        return
    }
    // Expect path like /api/workflows/{id}
    path := r.URL.Path
    // Trim prefix
    prefix := "/api/workflows/"
    if !strings.HasPrefix(path, prefix) {
        writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
        return
    }
    id := strings.TrimPrefix(path, prefix)
    if id == "" || strings.Contains(id, "/") {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
        return
    }
    // Return a DAG based on id (mock variants)
    wf := mockWorkflowByID(id)
    writeJSON(w, http.StatusOK, wf)
}
