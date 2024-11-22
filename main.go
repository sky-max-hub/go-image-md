package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var currentDir = "C:\\Users\\whither\\Documents\\goProjects\\RwhitherBlog\\content"
var suffix = ".md"
var pattern = "!\\[.*?\\]\\((https?://.*?)(?:\\s+\".*?\")?\\)"
var regex = regexp.MustCompile(pattern)
var savePath = "image"
var saveChannel = make(chan [2]string, 1000)
var goRoutineLimit = 20
var wg sync.WaitGroup

func main() {
	beginDownloadImage()
	// 遍历文件夹内的所有文件
	err := filepath.WalkDir(currentDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}
		// 如果是文件夹或者不存在指定后缀,不处理
		if d.IsDir() || !strings.HasSuffix(path, suffix) {
			return nil
		}
		// 打印文件或目录的名称
		fmt.Println("Visited:", path)
		// 读取md中符合image条件的内容
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Failed to read file %s: %v\n", path, err)
			return err
		}
		folderPath := ""
		createFolder := true
		replacedContent := regex.ReplaceAllStringFunc(string(content), func(match string) string {
			// 提取图片链接中的 URL（第一个捕获组）
			if createFolder {
				folderPath = createDir(path)
				createFolder = false
			}
			url := regex.FindStringSubmatch(match)[1]
			wg.Add(1)
			saveChannel <- [2]string{folderPath, url}
			// 在这里可以根据 URL 做不同的替换操作
			fileName := url[strings.LastIndex(url, "/")+1:]
			newURL := savePath + "/" + fileName
			// 返回替换后的 Markdown 图片链接
			return fmt.Sprintf(`![%s](%s)`, fileName, newURL)
		})
		// 将替换后的内容写回文件
		err = os.WriteFile(path, []byte(replacedContent), 0644)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path %q: %v\n", currentDir, err)
	}
	close(saveChannel) // 关闭通道
	wg.Wait()          // 等待消费者完成
}

func createDir(path string) string {
	folderPath := path[:strings.LastIndex(path, "\\")]
	folderPath = filepath.Join(folderPath, savePath)
	fmt.Println("Creating folder:", folderPath)
	// 创建文件夹，如果文件夹不存在
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err := os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			fmt.Printf("failed to create directory: %v\n", err)
		}
	}
	return folderPath
}

func beginDownloadImage() {
	for i := 0; i < goRoutineLimit; i++ {
		go func() {
			for {
				array, ok := <-saveChannel
				if !ok {
					return
				}
				folderPath, url := array[0], array[1]
				fmt.Println("Downloading", url, folderPath)
				fileName := url[strings.LastIndex(url, "/"):]
				filePath := filepath.Join(folderPath, fileName)
				fmt.Println("Creating file:", filePath)
				file, err := os.Create(filePath)
				defer file.Close()
				if err != nil {
					fmt.Printf("failed to create file: %v\n", err)
					return
				}
				// 发送 HTTP GET 请求下载图片
				resp, err := http.Get(url)
				if err != nil {
					fmt.Printf("failed to download image: %v\n", err)
					return
				}
				defer resp.Body.Close()
				// 将图片内容写入文件
				_, err = io.Copy(file, resp.Body)
				if err != nil {
					fmt.Printf("failed to save image: %v\n", err)
				}
				wg.Done()
			}
		}()
	}
}
