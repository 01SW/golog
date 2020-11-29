package golog

import (
	"encoding/json"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/ioutil"
	"os"
	"path"
	"time"
)

var logger *zap.Logger = nil

//日志输出方式
type LogWriteType int8
const (
	File LogWriteType = 0 //文件
	Stdout LogWriteType = 1 //终端
)

//日志配置
type logConfig struct{
	WriteType LogWriteType //日志输出方式，默认为File
	LevelEnum zapcore.Level //日志记录最低级别，默认为WarnLevel
	Encoding string //日志编码方式，默认为json
	SavePath string //日志保存路径，当WriteType为File时起效，默认为当前目录的log文件夹内的log.txt
	FileMaxSize int //日志文件最大数量(单位MB)，默认为10MB
	FileMaxBackups int //保留的旧日志文件最大数量， 默认为5个
	FileMaxAge int //保留日志最大日期，默认为30天
}

//日志初始化工具类
type News struct {
}

//获取初始配置数据
func (ns *News)GetDefaultConfig()*logConfig {
	defaultConfig := logConfig{}
	defaultConfig.WriteType = File
	defaultConfig.LevelEnum = zapcore.WarnLevel
	defaultConfig.Encoding = "json"
	defaultConfig.SavePath = "log/log.txt"
	defaultConfig.FileMaxSize = 10
	defaultConfig.FileMaxBackups = 5
	defaultConfig.FileMaxAge = 30

	return &defaultConfig
}

//获取指定了保存路径的配置数据(其余与初始配置一致)
func (ns *News)GetConfig(path string)*logConfig{
	config := ns.GetDefaultConfig()
	config.SavePath = path
	return config
}

//判断对应文件夹是否存在，如果不存在则创建
func mkDir(dirPath string){
	dir := path.Dir(dirPath)
	if _, er := os.Stat(dir); er != nil{
		er = os.MkdirAll(dir, os.ModePerm)
		if er != nil{
			panic("无法创建文件夹，" + er.Error())
		}
	}
}

//LogConfig转zap.Config
func configChangeZap(config *logConfig)*zap.Config{
	cfg := zap.Config{}

	var outType []string
	if config.WriteType == File{
		outType = append(outType, config.SavePath)
	}else{
		outType = append(outType, "stdout")
	}
	cfg.OutputPaths = outType
	cfg.Level = zap.NewAtomicLevelAt(config.LevelEnum)
	cfg.Encoding = config.Encoding
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.EncoderConfig.EncodeTime = timeEncoder

	mkDir(config.SavePath)

	return &cfg
}

//时间格式
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

//日志分割
func getLogWriter(config *logConfig)*lumberjack.Logger{
	lj := lumberjack.Logger{
		Filename:   config.SavePath,           // 输出文件
		MaxSize:    config.FileMaxSize,       // 日志文件最大大小（单位：MB）
		MaxBackups: config.FileMaxBackups,    // 保留的旧日志文件最大数量
		MaxAge:     config.FileMaxAge,        // 保存日期
	}
	return &lj
}

//初始化Log，需传递log配置文件路径，若该配置文件不存在则会以默认初始配置进行初始化
func (ns *News)Init(path string){
	var ler error
	var config *logConfig
	if _, er := os.Stat(path); er != nil{
		config = ns.GetDefaultConfig()
		cfg := configChangeZap(config)
		logger, ler = cfg.Build()
		if ler != nil{
			panic("初始化日志失败，" + ler.Error())
		}

		er = ns.SaveConfig(path, config)
		if er != nil{
			panic("无法保存配置文件，" + er.Error())
		}
	}else{
		file, er := os.Open(path)
		if er != nil{
			panic("无法打开配置文件，" + er.Error())
		}
		data, fer := ioutil.ReadAll(file)
		if fer != nil{
			panic("读取配置文件失败，" + fer.Error())
		}

		cer := json.Unmarshal(data, config)
		if cer != nil{
			panic("所读取的配置文件存在错误，" + cer.Error())
		}
		cfg := configChangeZap(config)
		logger, ler = cfg.Build()

		if ler != nil{
			panic("初始化日志失败，" + ler.Error())
		}
	}

	zap.ReplaceGlobals(logger)
	if config != nil && config.WriteType == File{
		zapcore.AddSync(getLogWriter(config))
	}
}

//以LogConfig结构体对象初始化日志
func (ns *News)ConfigInit(config *logConfig){
	cfg := configChangeZap(config)
	var er error
	logger, er = cfg.Build()
	if er != nil{
		panic("无法创建日志对象，" + er.Error())
	}
	zap.ReplaceGlobals(logger)
	if config != nil && config.WriteType == File{
		zapcore.AddSync(getLogWriter(config))
	}
}

//将LogConfig结构体对象保存到本地
func (ns *News)SaveConfig(path string, config *logConfig)error{
	mkDir(path)

	file, er := os.Create(path)
	if er != nil{
		return er
	}
	data, _ := json.MarshalIndent(config, "", " ")
	_, wer := file.Write(data)
	if wer != nil{
		return wer
	}
	_ = file.Close()
	return nil
}

//判断配置文件是否存在
func (ns *News)ConfigExist(path string)bool{
	if _, er := os.Stat(path); er != nil{
		return false
	}else{
		return true
	}
}

//获取Log单例指针，调用此函数前请先调用初始化函数
func GetLogger() *zap.Logger{
	if logger == nil{
		panic("未调用初始化函数")
	}

	return logger
}
