package util

const FinalizerName = "dns.xzzpig.com/finalizer"

func ContainsString(strings []string, str string) bool {
	if strings == nil {
		return false
	}
	for _, s := range strings {
		if s == str {
			return true
		}
	}
	return false
}

func RemoveString(strs []string, str string) []string {
	if len(strs) == 0 {
		return strs
	}
	index := 0
	for index < len(strs) {
		if strs[index] == str {
			strs = append(strs[:index], strs[index+1:]...)
			continue
		}
		index++
	}
	return strs
}
