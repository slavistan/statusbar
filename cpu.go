package main

type CpuConfig struct {}

//https://rosettacode.org/wiki/Linux_CPU_utilization#Go
// func main() {
// 	f, err := os.Open("/proc/stat")
// 	if err != nil {
// 		return
// 	}
// 	defer f.Close()

// 	var prefix string
// 	var a [7]float64
// 	if _, err = fmt.Fscanf(f, "%s %f %f %f %f %f %f %f", &prefix, &a[0], &a[1], &a[2], &a[3], &a[4], &a[5], &a[6]); err != nil {
// 		fmt.Print(err)
// 		return
// 	}
// 	sum := 0.0
// 	for _, v := range a {
// 		sum += v
// 	}

// 	fmt.Printf("%v\n", a)
// }
func (c CpuConfig) MakeStatusFn() StatusFn {
	return func(chan<- ModuleStatus) {

	}
}