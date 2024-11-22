# 批量扫描md图片，在线图片转本地图片

为什么会有这个需求？OSS不是很好，三天两头出问题，SSL证书过期，要么就是被别人刷流量。存在本地的话确实会好一点：

1. 不用担心博客图片突然消失了，很多博客文字还在，图片老早就不见了。。。
2. 不用担心被别人刷流量了，公共读的策略很方便，但是也方便了一些无聊的人刷流量。试过做一下防盗链，但是好像防君子不防小人。
3. 不用每天担心OSS的域名SSL过期了，阿里云的免费证书只能保持三个月，每次都要自己手动更新，很烦。
4. 确实信不过云上的OSS了，博客和图片分开存储总感觉不是太安全。

代码变量设置：

```go
// 设置的操作目录
var currentDir = "content"
// 处理的文件后缀
var suffix = ".md"
// 正则扫描的表达式
var pattern = "!\\[.*?\\]\\((https?://.*?)(?:\\s+\".*?\")?\\)"
var regex = regexp.MustCompile(pattern)
// 保存的图片路径，默认保存在每个md相对路径上的image文件夹
var savePath = "image"
var saveChannel = make(chan [2]string, 1000)
// 处理的携程限制，随便设置。。
var goRoutineLimit = 20
var wg sync.WaitGroup
```