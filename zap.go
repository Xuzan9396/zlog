package zlog


func Info(args ...interface{}) {
	F("sign","").Info( args...)
}


func  Error(args ...interface{}) {
	F("sign","").Error(args...)
}


// Debug uses fmt.Sprint to construct and log a message.
func  Debug(args ...interface{}) {
	F("sign","").Debug(args...)
}


// Warn uses fmt.Sprint to construct and log a message.
func  Warn(args ...interface{}) {
	F("sign","").Debug(args...)
}


// Panic uses fmt.Sprint to construct and log a message, then panics.
func Panic(args ...interface{}) {
	F("sign","").Panic(args...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func Fatal(args ...interface{}) {
	F("sign","").Fatal(args...)
}

func Infof(template string, args ...interface{}) {
	F("sign","").Infof(template , args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func  Warnf(template string, args ...interface{}) {
	F("sign","").Warnf(template , args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(template string, args ...interface{}) {
	F("sign","").Errorf(template , args...)

}


// Panicf uses fmt.Sprintf to log a templated message, then panics.
func  Panicf(template string, args ...interface{}) {
	F("sign","").Panicf(template , args...)

}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func  Fatalf(template string, args ...interface{}) {
	F("sign","").Fatalf(template , args...)
}
