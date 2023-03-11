package main

type functionInvocationCount struct {
	owner     string
	app       string
	function  string
	trigger   string
	perMinute [1441]int
	duration  int
	memory    int
}

type functionExecutionDuration struct {
	owner                string
	app                  string
	function             string
	average              int
	count                int
	minimum              int
	maximum              int
	percentileAverage0   int
	percentileAverage1   int
	percentileAverage25  int
	percentileAverage50  int
	percentileAverage75  int
	percentileAverage99  int
	percentileAverage100 int
}

type appMemory struct {
	owner                string
	app                  string
	count                int
	average              int
	percentileAverage1   int
	percentileAverage5   int
	percentileAverage25  int
	percentileAverage50  int
	percentileAverage75  int
	percentileAverage95  int
	percentileAverage99  int
	percentileAverage100 int
}
