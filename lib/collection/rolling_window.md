keywords: 滚动窗口、滑动窗口、断路器、容错、限速、fault tolerance、breaker、rate limit

## 一、滚窗算法(rolling window) [滚窗算法](https://medium.com/@anikdutta20031986/circuit-breaker-pattern-for-microservice-a2bb7213398e)
受 Netflix的 Hystrix 启发
![rolling window](https://miro.medium.com/max/576/1*jfOsUOv0uwTc4z-SbTMvuQ.png)

## 二、滑窗算法(sliding window) [滑窗算法](https://konghq.com/blog/how-to-design-a-scalable-rate-limiting-algorithm/) 
假设我们有一个限速器，60秒允许100个事件，现在时间到了 `75秒` 这个点，则滑动窗口如下所示：

![sliding window](https://github.com/RussellLuo/slidingwindow/raw/master/docs/slidingwindow.png)

该场景下，限速器在上一个窗口（灰色）已允许86个事件，当前窗口（橙色）已允许12个事件，那么，
滑动窗口的期间内，计数近似值为：

```
count = 86 * ((60-15)/60) + 12
      = 86 * 0.75 + 12
      = 64.5 + 12
      = 76.5   
```

---

**计算推理过程**

---

已知：
- 上一窗口已允许事件数86个，即：`prevPermitted = 86`
- 当前窗口已允许事件数12个，即：`currPermitted = 12`
- 窗口（含滑窗）时长相同，即：窗口时长 `duratin = 60`
- 滑窗与上窗的起始时间偏移为15秒，即：`offset = 15`
---

要求：
- 滑窗期内`可能的`事件个数 `slidingCount`

推理：
- 找出滑窗与上窗重叠的事件个数 `numOverlapWithPrevWin`
- 找出滑窗与当窗重叠的事件个数，已得知 `currPermitted`
- 将上窗重叠数与下窗重叠数相加，即得滑窗期内可能事件数

公式：
- 滑窗与上窗重叠比例：`ratio = (duration-offset)/duration`
- 滑窗与上窗重叠数量：`numOverlapWithPrevWin = prevPermitted * ratio`
- 滑窗期内可能事件数：`numOverlapWithPrevWin + currPermitted` 