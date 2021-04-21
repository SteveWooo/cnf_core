import numpy as np
import threading
import matplotlib.pyplot as mp
from mpl_toolkits.mplot3d import Axes3D
import json

def GetData() :
    fo = open("./avgMem.data", "r+")
    file = fo.read()
    fo.close()
    data = file.split("\n")
    return data[:len(data) - 1]

fontTitle = {"size":25}
fontXYZ = {"size":18}

def DrawCPU(data):
    X = []
    Y = []
    for index in range(len(data)):
        item = data[index].split(" ")
        X.append(float(item[0]))
        Y.append(float(item[1]))

    mp.plot(X, Y, linewidth=2.5, linestyle="-", label="Average CPU Usage (percentage)")
    # mp.plot(masterAreaKadX, masterAreaKadY, linewidth=2.5, linestyle="-", label="Area-Kad")
    mp.xlabel('Time (s)', fontdict=fontXYZ)
    mp.ylabel('CPU Usage (percentage)', fontdict=fontXYZ)

    # mp.xlim((0, 10000))
    mp.ylim((0, 100))

    mp.title("CPU Usage condition", fontdict=fontTitle)

    mp.legend(loc='upper left')
    return 

def DrawMem(data):
    X = []
    Y = []
    for index in range(len(data)):
        item = data[index].split(" ")
        X.append(float(item[0]))
        Y.append(float(item[2]))

    mp.plot(X, Y, linewidth=2.5, linestyle="-", label="Average Memory Usage (Mb)")
    # mp.plot(masterAreaKadX, masterAreaKadY, linewidth=2.5, linestyle="-", label="Area-Kad")
    mp.xlabel('Time (s)', fontdict=fontXYZ)
    mp.ylabel('Memory Usage (Mb)', fontdict=fontXYZ)

    # mp.xlim((0, 10000))
    mp.ylim((0, 3200))

    mp.title("Memory Usage condition", fontdict=fontTitle)

    mp.legend(loc='upper left')
    return 

def DrawNode(data):
    X = []
    Y = []
    for index in range(len(data)):
        item = data[index].split(" ")
        X.append(float(item[0]))
        Y.append(float(item[3]))

    mp.plot(X, Y, linewidth=2.5, linestyle="-", label="Node Count")
    # mp.plot(masterAreaKadX, masterAreaKadY, linewidth=2.5, linestyle="-", label="Area-Kad")
    mp.xlabel('Time (s)', fontdict=fontXYZ)
    mp.ylabel('Node Count', fontdict=fontXYZ)

    # mp.xlim((0, 10000))
    mp.ylim((0, 70000))

    mp.title("Active Node Count", fontdict=fontTitle)

    mp.legend(loc='upper left')
    return 

def DrawConnCount(data):
    X = []
    Y = []
    for index in range(len(data)):
        item = data[index].split(" ")
        X.append(float(item[0]))
        Y.append(float(item[4]))

    mp.plot(X, Y, linewidth=2.5, linestyle="-", label="Node Connection Count")
    # mp.plot(masterAreaKadX, masterAreaKadY, linewidth=2.5, linestyle="-", label="Area-Kad")
    mp.xlabel('Time (s)', fontdict=fontXYZ)
    mp.ylabel('Connection Count', fontdict=fontXYZ)

    # mp.xlim((0, 10000))
    # mp.ylim((0, 700000))

    mp.title("Valid Connection Count", fontdict=fontTitle)

    mp.legend(loc='upper left')
    return 

data = GetData()

mp.figure("系统运行状况")

mp.subplot(2, 2, 1)
DrawCPU(data)

mp.subplot(2, 2, 2)
DrawMem(data)

mp.subplot(2, 2, 3)
DrawNode(data)

mp.subplot(2, 2, 4)
DrawConnCount(data)

mp.show()