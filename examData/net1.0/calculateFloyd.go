package main

import (
	json "encoding/json"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/cnf_core/src/utils/logger"
)

func main() {
	filePath := "./" + os.Args[1] + "/" + os.Args[2]
	sourceFile, _ := ioutil.ReadFile(filePath + "/originFloydMat.json")
	fileString := string(sourceFile)
	var fileJSON []interface{}
	decodeErr := json.Unmarshal([]byte(fileString), &fileJSON)
	if decodeErr != nil {
	}

	var mat [][]int
	var line []interface{}
	for i := 0; i < len(fileJSON); i++ {
		line = fileJSON[i].([]interface{})
		lineMat := make([]int, len(line))
		for k := 0; k < len(line); k++ {
			if line[k] == "max" {
				lineMat[k] = -1
			}

			if line[k] == "1" {
				lineMat[k] = 1
			}

			if line[k] == "0" {
				lineMat[k] = 0
			}
		}

		mat = append(mat, lineMat)
	}

	// logger.Debug("开始计算")

	for i := 0; i < len(mat); i++ {
		for k := 0; k < len(mat); k++ {
			for j := 0; j < len(mat); j++ {
				if i == k || i == j || j == k {
					continue
				}
				kjLength := mat[k][j]
				kijLength := -1
				// 绕路的两条路都不是无穷大的话
				if mat[k][i] != -1 && mat[i][j] != -1 {
					kijLength = mat[k][i] + mat[i][j]
				}

				// 如果绕路走是无穷大 就不管
				if kijLength == -1 {
					continue
				}

				// 如果绕路走不是无穷大，而直接走是无穷大
				if kijLength != -1 && kjLength == -1 {
					mat[k][j] = kijLength
					continue
				}

				if kijLength != -1 && kjLength != -1 {
					if kijLength < kjLength {
						mat[k][j] = kijLength
					}
					continue
				}
			}
		}
		// logger.Debug("done : " + strconv.Itoa(i) + "/" + strconv.Itoa(len(mat)))
	}

	matJSON, _ := json.Marshal(mat)
	ioutil.WriteFile(filePath+"/floydMat.json", matJSON, 0644)

	var sumDistance int = 0
	var maxDistance int = 0
	var nodeCount int = len(mat)

	var lineCount float64 = float64(nodeCount) * float64(nodeCount-1) // 作为平均数的分母

	for i := 0; i < len(mat); i++ {
		for k := 0; k < len(mat); k++ {
			if mat[i][k] > maxDistance {
				maxDistance = mat[i][k]
			}

			sumDistance += mat[i][k]
		}
	}

	var avgConn float64 = float64(sumDistance) / lineCount

	logger.Debug("算法：" + os.Args[1])
	logger.Debug("结点数：" + strconv.Itoa(nodeCount))
	logger.Debug("平均节点间距离：" + strconv.FormatFloat(avgConn, 'E', -1, 64))
	logger.Debug("最大节点间距离：" + strconv.Itoa(maxDistance))

}
