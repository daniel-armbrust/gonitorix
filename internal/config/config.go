package config

type GlobalConfig struct {
	RRDPath string `yaml:"rrd_path"`
	ImgPath string `yaml:"img_path"`
}

type Config struct {
	Global GlobalConfig
	NetIf  NetIfConfig
}