package main

import (
    "net/http"
)

type WorkflowNode struct {
    ID     string  `json:"id"`
    Name   string  `json:"name"`
    X      float64 `json:"x"`
    Y      float64 `json:"y"`
    Status string  `json:"status"`
    Desc   string  `json:"desc"`
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

// mock workflow data (same layout as the current Vue demo)
func mockWorkflow() WorkflowResponse {
    nodes := []WorkflowNode{
        {ID: "A", Name: "任务A", X: 390, Y: 40, Status: "success", Desc: "数据准备"},
        {ID: "B", Name: "任务B", X: 250, Y: 120, Status: "running", Desc: "数据清洗"},
        {ID: "C", Name: "任务C", X: 530, Y: 120, Status: "success", Desc: "特征工程"},
        {ID: "D", Name: "任务D", X: 170, Y: 220, Status: "pending", Desc: "模型训练1"},
        {ID: "E", Name: "任务E", X: 310, Y: 220, Status: "pending", Desc: "模型训练2"},
        {ID: "F", Name: "任务F", X: 470, Y: 220, Status: "failed", Desc: "模型训练3"},
        {ID: "G", Name: "任务G", X: 610, Y: 220, Status: "pending", Desc: "模型训练4"},
        {ID: "H", Name: "任务H", X: 250, Y: 320, Status: "success", Desc: "评估1"},
        {ID: "I", Name: "任务I", X: 530, Y: 320, Status: "pending", Desc: "评估2"},
        {ID: "J", Name: "任务J", X: 390, Y: 400, Status: "pending", Desc: "结果汇总"},
        {ID: "K", Name: "任务K", X: 110, Y: 120, Status: "pending", Desc: "数据采样"},
        {ID: "L", Name: "任务L", X: 670, Y: 120, Status: "pending", Desc: "规则生成"},
        {ID: "M", Name: "任务M", X: 110, Y: 320, Status: "pending", Desc: "可视化报告"},
        {ID: "N", Name: "任务N", X: 670, Y: 320, Status: "pending", Desc: "在线部署"},
        {ID: "O", Name: "任务O", X: 30,  Y: 220, Status: "pending", Desc: "数据质量校验"},
        {ID: "P", Name: "任务P", X: 750, Y: 220, Status: "pending", Desc: "特征选择"},
        {ID: "Q", Name: "任务Q", X: 390, Y: 500, Status: "pending", Desc: "超参搜索"},
        {ID: "R", Name: "任务R", X: 250, Y: 580, Status: "pending", Desc: "A/B 测试"},
        {ID: "S", Name: "任务S", X: 530, Y: 580, Status: "pending", Desc: "回滚策略"},
        {ID: "T", Name: "任务T", X: 390, Y: 660, Status: "pending", Desc: "上线审核"},
        {ID: "U", Name: "任务U", X: 390, Y: 740, Status: "pending", Desc: "告警与监控"},
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
