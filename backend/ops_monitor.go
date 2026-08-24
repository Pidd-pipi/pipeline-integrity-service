package main

// opsLatencyStats 记录每个请求路径的耗时样本。
var opsLatencyStats = map[string][]int64{}

// opsLatencyRecord 追加一条耗时样本。
func opsLatencyRecord(path string, ms int64) {
	opsLatencyStats[path] = append(opsLatencyStats[path], ms)
}

// opsLatencySummary 返回指定路径的样本数与平均耗时。
func opsLatencySummary(path string) (count int, avg int64) {
	samples := opsLatencyStats[path]
	if len(samples) == 0 {
		return 0, 0
	}
	var total int64
	for _, v := range samples {
		total += v
	}
	return len(samples), total / int64(len(samples))
}

// opsLatencyReset 清空全部样本。
func opsLatencyReset() {
	opsLatencyStats = map[string][]int64{}
}
