const Express = require("express")
const Cluster = require("cluster")
const Dgram = require("dgram")
const fs = require("fs")
const request = require('request')

/**
 * node --max-old-space-size=16384 startup.js // 拿满16G内存
 */

var masterStatus = {
    nodes : {},
    updateNodes : async function(jsonData){
        // console.log(jsonData)
        if (jsonData["nodeID"] == undefined) {
            return 
        }
        masterStatus.nodes[jsonData["nodeID"]] = jsonData
    },
    // 搞出一个floyd算法
    lastNodeStatus : {},
    // 结点连接情况稳定次数
    doneTime : 0,
    // 结点连接情况稳定上限，到达这个上限，就写文件，走下一步
    ALL_DONE_TIME : 4,

    // 是否使用区域Master算法，这些参数需要作为环境变量传递下去。mats_masterAreaKad代表使用区域master算法
    // 
    logDirName : "mats_originKad",
    // logDirName : "mats_masterAreaKad",
    nodeCount : 2800, // 初始结点数
    nodeCountMax : 2800, // 最大结点数

    calculate : async function() {
        // 当大家的对外连接都稳定有8个了，就写状态
        var done = true
        var length = 0
        var node
        for(var nodeID in masterStatus.nodes) {
            node = masterStatus.nodes[nodeID]
            if (masterStatus.lastNodeStatus[nodeID] == undefined) {
                masterStatus.lastNodeStatus[nodeID] = 0
            }

            length = 0
            if (node.netStatus.nodeConnectionStatus.serviceStatus.outBoundConn != undefined) {
                length += node.netStatus.nodeConnectionStatus.serviceStatus.outBoundConn.length
            }
            if (node.netStatus.nodeConnectionStatus.serviceStatus.inBoundConn != undefined) {
                length += node.netStatus.nodeConnectionStatus.serviceStatus.inBoundConn.length
            }

            // 对比近两次的状态，如果都没变化，就写
            if (length != masterStatus.lastNodeStatus[nodeID]) {
                done = false
            }

            masterStatus.lastNodeStatus[nodeID] = length
        }
        if (done) {
            masterStatus.doneTime ++
            if (masterStatus.doneTime > masterStatus.ALL_DONE_TIME) {
                return 
            }
            // console.log(`done time: ${masterStatus.doneTime}`)
        } else {
            masterStatus.doneTime = 0
        }

        if (masterStatus.doneTime == masterStatus.ALL_DONE_TIME) {
            // 数一下结点数
            let nodeCount = Object.keys(masterStatus.nodes).length
            try{
                fs.mkdirSync(`${__dirname}/${masterStatus.logDirName}/${nodeCount}`)
            }catch(e) {

            }
            fs.writeFileSync(`${__dirname}/${masterStatus.logDirName}/${nodeCount}/originMat.json`, JSON.stringify(masterStatus.nodes))
            if(process.argv[2] != "onlyshow") {
                console.log(`数据采集完成，开始计算`)
                try{
                    var stdout = require('child_process').execSync(`node --max-old-space-size=16384 ${__dirname}/translateToFloydMat.js ${masterStatus.logDirName} ${nodeCount}`)
                    console.log(stdout.toString())
                    process.send(1)
                } catch(e) {
                    console.log(e)
                }
            }   
        }
        
        return 
    },

    storeLogData : async function(time){
        let count = 0;
        let totalMem = 0 // 占用系统内存，单位MB
        let totalCPUUsage = 0 // CPU消耗百分比
        let node;
        let outboundLength = 0
        for(var nodeID in masterStatus.nodes) {
            node = masterStatus.nodes[nodeID]
            count ++ 
            totalMem += node.osMem
            totalCPUUsage += node.osCPUUsage
            // 对外连接总数
            if (node.netStatus.nodeConnectionStatus.serviceStatus.outBoundConn != undefined) {
                outboundLength += node.netStatus.nodeConnectionStatus.serviceStatus.outBoundConn.length
            }
            
        }

        console.log(`Avg mem: ${totalMem / count}, Avg cpu usage: ${totalCPUUsage / count}, node count: ${count} , total outbound connection: ${outboundLength}`)
        if (count == 0) {
            return 
        }
        if(process.argv[2] != "onlyshow") {
            fs.appendFileSync(`${__dirname}/avgMem.data`, `${time} ${totalCPUUsage / count} ${totalMem / count} ${count} ${outboundLength}\n`)
        }
        
    }
}

let masterJob = {
    httpServer : {
        init : async function(){
            let app = Express();

            app.use("/public", Express.static(`${__dirname}/public`));

            app.get("/api/node_status", async function(req, res) {
                // 清掉没有心跳的节点
                let now = +new Date();
                // for(var nodeId in masterStatus.nodes) {
                //     if (now - masterStatus.nodes[nodeId].update >= 5000) {
                //         masterStatus.nodes[nodeId].nodeStatus = 'disconnect';
                //     }
                // }

                res.send(JSON.stringify({
                    status : 2000,
                    nodes : masterStatus.nodes
                }));
            })

            app.post("/api/update_node_status", async function(req, res) {
                var body = '', jsonStr;
                req.on('data', async function (chunk) {
                    body += chunk; //读取参数流转化为字符串
                });
                req.on('end', async function () {
                    //读取参数流结束后将转化的body字符串解析成 JSON 格式
                    try {
                        jsonStr = JSON.parse(body);

                        for(var nodeID in jsonStr.nodes) {
                            jsonStr.nodes[nodeID].osCPUUsage = jsonStr["osCPUUsage"]
                            jsonStr.nodes[nodeID].osMem = jsonStr['osMem']
                            await masterStatus.updateNodes(jsonStr.nodes[nodeID])
                        }

                        res.send(JSON.stringify({
                            status : 2000,
                        }));
                    } catch (err) {
                        res.send(JSON.stringify({
                            status : 4000,
                        }));
                        jsonStr = null;
                    }
                })
            })

            let port = 8081;
            app.listen(port, function(){
                // console.log("app listen : " + port)
            })

            // udp服务器
            let udpSoceket = Dgram.createSocket("udp4")
            udpSoceket.on("listening", async function(){
                // console.log("udp server listen at : " + port)
            })
    
            udpSoceket.on("message", async function(message, remote){
                message = message.toString().split("\r\n")[0]
                let data ;
                try{
                    data = JSON.parse(message)
                    console.log(data)
                }catch(e) {
                    console.log(e)
                }

                await masterStatus.updateNodes(data)
            })
    
            udpSoceket.bind({
                exclusive : true,
                address : "127.0.0.1",
                port : port
            })
        }
    }
}

async function main(){
    var nodeWorker, goWorker
    if(Cluster.isMaster) {
        setInterval(async function(){
            if (nodeWorker == undefined && goWorker == undefined) {
                if (masterStatus.nodeCount > masterStatus.nodeCountMax) {
                    console.log("所有轮次实验已完成，请退出主程序")
                    process.exit(0)
                    return 
                }
                console.log(`==================================`)
                // 子进程环境变量
                var forkEnv = {
                    "RUNING_MODULE" : "", // 下面填
                    "CNF_NODECOUNT" : masterStatus.nodeCount,
                    "LOG_DIR_NAME" : masterStatus.logDirName
                }

                forkEnv["RUNING_MODULE"] = "nodecnf"
                nodeWorker = Cluster.fork(forkEnv)
                nodeWorker.on("exit", async function(){
                    // console.log("node cnf on exit")
                    nodeWorker = undefined
                })

                // 只演示的话，不需要启动go进程
                if(process.argv[2] != "onlyshow") {
                    forkEnv["RUNING_MODULE"] = "gocnf"
                    goWorker = Cluster.fork(forkEnv)
                    goWorker.on("exit", async function(){
                        // console.log("go cnf on exit")
                        goWorker = undefined
                    })
                }

                nodeWorker.on("message", async function(msg) {
                    // 1 代表正常退出，可以开始进行下一轮实验
                    if(msg == 1) {
                        // 如果这个是使用originkad算法，就转换成masterArea算法，直接开始下一轮
                        if (masterStatus.logDirName == "mats_originKad") {
                            masterStatus.logDirName = "mats_masterAreaKad"
                        } else { // 如果这个已经使用了masterAreaKad算法，那就加一点结点再开始一轮
                            masterStatus.logDirName = "mats_originKad"
                            masterStatus.nodeCount += 100
                        }

                        // 只演示的话，不需要杀掉进程，让他自己一直跑下去就行了
                        if(process.argv[2] != "onlyshow") {
                            nodeWorker.send("kill")
                            goWorker.send("kill")
                        }
                    }
                })
            }
        }, 4000)

    } else {
        // 设置统一变量
        masterStatus.nodeCount = process.env["CNF_NODECOUNT"]
        masterStatus.logDirName = process.env["LOG_DIR_NAME"]
        // 判断启动的环境变量
        if (process.env["RUNING_MODULE"] == "nodecnf") {
            process.on("message", function(msg) {
                // console.log(`node worker receive : ${msg}`)
                if(msg == "kill") {
                    process.exit(0)
                }
            })
            await masterJob.httpServer.init();

            // 防止上一次没算完，就进入下一次计算
            // var doneCalculate = true
            // setInterval(async function(){
            //     // 没算完就退出
            //     if (doneCalculate == false ) {
            //         return 
            //     }

            //     doneCalculate = false
            //     await masterStatus.calculate()
            //     doneCalculate = true
            // }, 3000)

            // 定时保存一份日志
            let time = 0
            let doneStorage = true
            setInterval(async function(){
                if (doneStorage == true) {
                    doneStorage = false
                    await masterStatus.storeLogData(time)
                    time ++ 
                    doneStorage = true
                }
            }, 3000)

        }

        if (process.env["RUNING_MODULE"] == "gocnf") {
            console.log(`${new Date()} 开始实验 \n结点数:${masterStatus.nodeCount} 使用算法:${masterStatus.logDirName}`)
            process.on("message", function(msg) {
                if(msg == "kill") {
                    // 使用http请求，让go进程自己结束，不然会占用端口。
                    request("http://localhost:8082/exit", async function(err, res, body) {
                        process.exit(0)
                    })
                }
            })
            // 启动go进程
            require('child_process').exec(`run.bat ${masterStatus.nodeCount} ${masterStatus.logDirName}`, {
                "cwd" : `${__dirname}/../../../cnf_core`
            }, function(err, stdout){
                if (err) {
                    console.log(err)
                }
            })
        }
    }
}
main();