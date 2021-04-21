const fs = require('fs')
async function main(){
    let nodes = require(`${__dirname}/${process.argv[2]}/${process.argv[3]}/originMat.json`)
    var nodeIDIndex = {}
    var mat = []
    var path = {}
    var nodeCount = 0

    // 映射NodeID到矩阵
    for (var nodeID in nodes) {
        nodeIDIndex[nodeID] = nodeCount
        nodeCount ++;
    }

    // 初始化矩阵
    for(var nodeID in nodes) {
        // mat[nodeID] = {}
        mat[nodeIDIndex[nodeID]] = []
        path[nodeID] = {}
        for (var subNodeID in nodes) {
            if (subNodeID == nodeID) {
                mat[nodeIDIndex[nodeID]][nodeIDIndex[subNodeID]] = "0"
                continue
            }
            path[nodeID][subNodeID] = undefined // 最短路径的必经路径
            mat[nodeIDIndex[nodeID]][nodeIDIndex[subNodeID]] = "max" // 初始化两个路径正无穷远
        }
    }
    // 初始化各个结点之间的连接
    for(var nodeID in nodes) {
        var node = nodes[nodeID]
        var conn = node.netStatus.nodeConnectionStatus.serviceStatus
        if (conn.inBoundConn != null) {
            for (var i=0;i<conn.inBoundConn.length;i++) {
                mat[nodeIDIndex[nodeID]][nodeIDIndex[conn.inBoundConn[i]]] = "1"
            }
        }
        if (conn.outBoundConn != null) {
            for (var i=0;i<conn.outBoundConn.length;i++) {
                mat[nodeIDIndex[nodeID]][nodeIDIndex[conn.outBoundConn[i]]] = "1"
            }
        }
    }

    fs.writeFileSync(`${__dirname}/${process.argv[2]}/${process.argv[3]}/originFloydMat.json`, JSON.stringify(mat))
    // console.log("转换完成")
    try{
        var stdout = require('child_process').execSync(`go run ${__dirname}\\calculateFloyd.go ${process.argv[2]} ${process.argv[3]}`)
        console.log(stdout.toString())
    }catch(e) {
        console.log(e)
    }
    return 
    console.log("开始计算")
    for (var i=0;i<mat.length;i++) {
        console.log(`完成：${i} / ${mat.length}`)
        for(var k=0;k<mat.length;k++) {
            for(var j=0;j<mat.length;j++) {
                if (i == 0 || i == k || j == k) {
                    continue
                }

                if (mat[j][k] > mat[j][i] + mat[i][k]) {
                    mat[j][k] = mat[j][i] + mat[i][k]
                }
            }
        }
    }

    fs.writeFileSync(`${__dirname}/floydMat.json`, JSON.stringify(mat))
}
main();