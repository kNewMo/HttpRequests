// 文件类，扩展支持文件创建时间，才只测试过windows，linux
// 对interface和struct还是没有很了解，但是下面的代码可以用了，效率，内存不确定

package file

import (
	"os"
	"strings"
)

const PathSeparator string = string(os.PathSeparator)
var FileMode os.FileMode = 0644

// 获取程序所在目录
func Root() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	return path[0:strings.LastIndex(path, PathSeparator) + 1], nil
}

// 创建一个文件
func Create(name string) (*os.File, error) {
	 return os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, FileMode)
}

// 判断文件是否存在
func Exist(name string) (bool) {
	_, err := os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}