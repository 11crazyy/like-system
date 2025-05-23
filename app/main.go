package main

import (
	"flag"
	"fmt"
	"index/models"
	"io"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"github.com/sirupsen/logrus"
)

var GitCommit string
var BuildTime string

func main() {
	var addr string = carrot.GetEnv("ADDR")
	var logFile string = carrot.GetEnv("LOG_FILE")
	var runMigration bool
	var logerLevel string = carrot.GetEnv("LOG_LEVEL")
	var dbDriver string = carrot.GetEnv(carrot.ENV_DB_DRIVER)
	var dsn string = carrot.GetEnv(carrot.ENV_DSN)
	var traceSql bool = carrot.GetEnv("TRACE_SQL") != ""

	var superUserEmail string
	var superUserPassword string

	if addr == "" {
		addr = ":8002"
	}
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			return "", fmt.Sprintf("%s:%d", f.File, f.Line)
		},
	})

	flag.StringVar(&superUserEmail, "superuser", "", "Create an super user with email")
	flag.StringVar(&superUserPassword, "password", "", "Super user password")
	flag.StringVar(&addr, "addr", addr, "HTTP Serve address")
	flag.StringVar(&logFile, "log", logFile, "Log output file name, default is os.Stdout")
	flag.StringVar(&logerLevel, "level", logerLevel, "Log level debug|info|warn|error")
	flag.BoolVar(&runMigration, "m", false, "Run migration only")
	flag.StringVar(&dbDriver, "db", dbDriver, "DB Driver, sqlite|mysql")
	flag.StringVar(&dsn, "dsn", dsn, "DB DSN")
	flag.BoolVar(&traceSql, "tracesql", traceSql, "Trace sql execution")
	flag.Parse()

	var lw io.Writer = os.Stdout
	var err error

	if logFile != "" {
		lw, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("open %s fail, %v\n", logFile, err)
		} else {
			logrus.SetOutput(lw)
		}
	} else {
		logFile = "console"
	}

	if logerLevel != "" {
		level, err := logrus.ParseLevel(logerLevel)
		if err == nil {
			logrus.SetLevel(level)
		}
	}

	fmt.Println("GitCommit   =", GitCommit)
	fmt.Println("BuildTime   =", BuildTime)
	fmt.Println("Addr        =", addr)
	fmt.Println("Logfile     =", logFile)
	fmt.Println("LogerLevel  =", logerLevel)
	fmt.Println("DB Driver   =", dbDriver)
	fmt.Println("DSN         =", dsn)
	fmt.Println("TraceSql    =", traceSql)
	fmt.Println("Migration   =", runMigration)

	db, err := carrot.InitDatabase(lw, dbDriver, dsn)
	if err != nil {
		panic(err)
	}

	if traceSql {
		db = db.Debug()
	}

	if err = carrot.InitMigrate(db); err != nil {
		panic(err)
	}

	if err = models.Migration(db); err != nil {
		panic(err)
	}

	models.CheckDefaultValues(db)
	// Init Database
	if runMigration {
		fmt.Println("migration done 🎉")
		return
	}

	r := gin.New()

	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: lw,
	}), gin.Recovery())

	fmt.Println("🎉 Voicefox acd is running on", addr)

	//TODO 创建并启动同步任务 handler.Start()

	r.Run(addr)
}
