const fs = require('fs');

// 根据最终的floyd图，统计成excel图
async function main(){
    let dirs = fs.readdirSync(`${__dirname}/mats_originKad`)
    let result = []

    // 先对dir排序
    for(var i=0;i<dirs.length;i++) {
        for(var k=i+1;k<dirs.length;k++) {
            if(parseInt(dirs[i]) > parseInt(dirs[k])) {
                temp = dirs[i];
                dirs[i] = dirs[k];
                dirs[k] = temp
            }
        }
    }

    for(var i=0;i<dirs.length;i++) {
        var data = await handleDir(dirs[i])
        result.push(data)
    }
    handleResult(result)
}
main();

async function handleResult(result) {
    let nodeCount = ""
    let originAvgResult = ""
    let masterAreaAvgResult = ""
    let originMaxResult = ""
    let masterAreaMaxResult = ""

    // python读的JSON数据，一个数据一张图
    let pythonData = {
        avgShortPath : { // 平均最短路径图
            origin : {
                X : [],
                Y : []
            },
            masterArea : {
                X : [],
                Y : []
            },
        }, 
        maxPath : { // 最长路径比较图
            origin : {
                X : [],
                Y : []
            },
            masterArea : {
                X : [],
                Y : []
            },
        }, 

        origin3dData : {}, // originKad顶点分布图
        masterArea3dData : {}, // masterAreaKad顶点分布图
    }

    for(var i=0;i<result.length;i++) {
        nodeCount += `${result[i].origin.nodeCount}\n`
        // 统一横坐标
        pythonData.avgShortPath.origin.X.push(result[i].origin.nodeCount)
        pythonData.avgShortPath.masterArea.X.push(result[i].masterArea.nodeCount)
        pythonData.maxPath.origin.X.push(result[i].origin.nodeCount)
        pythonData.maxPath.masterArea.X.push(result[i].masterArea.nodeCount)
    }

    // 全部数据中的最大度数：
    let originMaxLineCount = 0
    let masterAreaMaxLineCount = 0
    for(var i=0;i<result.length;i++) {
        originAvgResult += `${result[i].origin.avgDistance}\n`
        pythonData.avgShortPath.origin.Y.push(result[i].origin.avgDistance)
        originMaxResult += `${result[i].origin.maxDistance}\n`
        pythonData.maxPath.origin.Y.push(result[i].origin.maxDistance)
        if (result[i].origin.maxDistance > originMaxLineCount) {
            originMaxLineCount = result[i].origin.maxDistance
        }

        masterAreaAvgResult += `${result[i].masterArea.avgDistance}\n`
        pythonData.avgShortPath.masterArea.Y.push(result[i].masterArea.avgDistance)
        masterAreaMaxResult += `${result[i].masterArea.maxDistance}\n`
        pythonData.maxPath.masterArea.Y.push(result[i].masterArea.maxDistance)
        if (result[i].masterArea.maxDistance > masterAreaMaxLineCount) {
            masterAreaMaxLineCount = result[i].masterArea.maxDistance
        }
    }

    let file = `nodeCount:\n${nodeCount}\n\noriginAvg:\n${originAvgResult}\n\noriginMax:\n${originMaxResult}\n\nmasterAreaAvg:\n${masterAreaAvgResult}\n\nmasterAreaMax:\n${masterAreaMaxResult}\n`;

    // 画三维数据：
    file += `originKad度数量分布：\norigin最大路径：${originMaxLineCount}, 数据组数：${result.length}\n`
    let origin3dData = {
        X : [], // 结点数
        Y : [], // 度数
        Z : []
    }
    for(var i=0;i<result.length;i++) {
        origin3dData.X.push(result[i].origin.nodeCount)
    }
    for(var i=0;i<originMaxLineCount;i++) {
        origin3dData.Y.push(i)
    }

    for(var i=0;i<originMaxLineCount;i++) {
        origin3dData.Z[i] = []
        for(var r=0;r<result.length;r++) {
            if (result[r].origin.shortPathDistribute[i] == undefined) {
                result[r].origin.shortPathDistribute[i] = 0
            }
            origin3dData.Z[i].push(result[r].origin.shortPathDistribute[i])
        }
    }

    file += `${JSON.stringify(origin3dData)}`

    // 画三维数据（masterArea
    file += `masterAreaKad度数量分布：\norigin最大路径：${masterAreaMaxLineCount}, 数据组数：${result.length}\n`
    let masterArea3dData = {
        X : [], // 结点数
        Y : [], // 度数
        Z : []
    }
    for(var i=0;i<result.length;i++) {
        masterArea3dData.X.push(result[i].masterArea.nodeCount)
    }
    for(var i=0;i<masterAreaMaxLineCount;i++) {
        masterArea3dData.Y.push(i)
    }

    for(var i=0;i<masterAreaMaxLineCount;i++) {
        masterArea3dData.Z[i] = []
        for(var r=0;r<result.length;r++) {
            if (result[r].masterArea.shortPathDistribute[i] == undefined) {
                result[r].masterArea.shortPathDistribute[i] = 0
            }
            masterArea3dData.Z[i].push(result[r].masterArea.shortPathDistribute[i])
        }
    }

    file += `${JSON.stringify(masterArea3dData)}`

    // fs.writeFileSync(`${__dirname}/filaldata.data`, file)

    // 写写JSON格式的数据，给python直接画图
    pythonData.origin3dData = origin3dData
    pythonData.masterArea3dData = masterArea3dData

    fs.writeFileSync(`${__dirname}/finalData.json`, JSON.stringify(pythonData))

}

async function handleDir(nodeCount) {
    let originKadMat = require(`${__dirname}/mats_originKad/${nodeCount}/floydMat.json`);
    let originKadMatData = await calculateMatData(originKadMat)
    // console.log(originKadMatData)

    let masterAreaKadMat = require(`${__dirname}/mats_masterAreaKad/${nodeCount}/floydMat.json`);
    let masterAreaKadMatData = await calculateMatData(masterAreaKadMat)
    // console.log(masterAreaKadMatData)

    return {
        origin: originKadMatData,
        masterArea : masterAreaKadMatData
    }
}

async function calculateMatData(mat) {
    // 平均节点间距离
    var sumDistance = 0
    // 最大节点间距离
    var maxDistance = 0
    var nodeCount = mat.length

    // 最短路径分布
    var shortPathDistribute = []

    for (var i=0;i<mat.length;i++) {
        for(var k=i+1;k<mat.length;k++) {
            if (mat[i][k] > maxDistance) {
                maxDistance = mat[i][k]
            }

            if (shortPathDistribute[mat[i][k]] == undefined) {
                shortPathDistribute[mat[i][k]] = 0
            }
            shortPathDistribute[mat[i][k]] ++ 

            sumDistance += mat[i][k]
        }
    }

    // console.log(shortPathDistribute)

    for(var i=0;i<shortPathDistribute.length;i++) {
        if (shortPathDistribute[i] == undefined) {
            shortPathDistribute[i] = 0
        }
    }

    return {
        avgDistance : (sumDistance * 2) / (nodeCount * (nodeCount - 1)),
        maxDistance : maxDistance,
        nodeCount : nodeCount,
        shortPathDistribute : shortPathDistribute
    }
}
// async function main(){
//     var mat = require(`${__dirname}/mats/originFloydMat.json`)
    
//     console.log("开始计算")
//     for (var i=0;i<mat.length;i++) {
//         console.log(`完成：${i} / ${mat.length}`)
//         for(var k=0;k<mat.length;k++) {
//             if (mat[i][k] == "max") {
//                 mat[i][k] = Infinity
//             } else {
//                 mat[i][k] = parseInt(mat[i][k])
//             }
//             for(var j=0;j<mat.length;j++) {
//                 if (i == 0 || i == k || j == k) {
//                     continue
//                 }

//                 if (mat[j][k] > mat[j][i] + mat[i][k]) {
//                     mat[j][k] = mat[j][i] + mat[i][k]
//                 }
//             }
//         }
//     }

//     // 平均节点间距离
//     var sumDistance = 0
//     var maxDistance = 0
//     var nodeCount = mat.length

//     for (var i=0;i<mat.length;i++) {
//         for(var k=0;k<mat.length;k++) {
//             if (mat[i][k] > maxDistance) {
//                 maxDistance = mat[i][k]
//             }

//             sumDistance += mat[i][k]
//         }
//     }

//     console.log(`============= report ============
//  结点数：${nodeCount}
//  平均节点间距离：${sumDistance / (nodeCount * nodeCount)}
//  最大节点间距离：${maxDistance}
//  `)
// }
// main();