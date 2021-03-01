# dubbo-go-proxy-tool
Dubbo Go 网关代理小工具

通过在扫描service目录下provider服务方法，生成网关代理方法配置, 例如在service/hello/目录下添加HelloProvider, 在项目根目录执行该工具后扫描service目录，并会匹配到带有@desc 与@used的方法，
生成服务方法配置到doc/config.go
```
package docs

var methodDocs = `[
  {
    "service_name": "nkobase.HelloProvider.SayHello",
    "interface_url": "la.kaike.nkobase.hello.HelloProvider",
    "params": [
      {
        "name": "string"
      }
    ],
    "group": "",
    "used_app_name": "test",
    "desc": "测试注释"
  }
]`
```

```
package hello

import (
	"context"
	"fmt"
	"github.com/apache/dubbo-go/common/constant"
	"github.com/apache/dubbo-go/config"
	"nko-utils/common"
	"nko-utils/mq"
)

func init() {
	config.SetProviderService(new(HelloProvider))
}

type HelloProvider struct {
}

//设置服务引用名 注意必须唯一不能重复
func (h *HelloProvider) Reference() string {
	return common.GetStructName(*h)
}

//@desc 测试注释
//@used test
func (h *HelloProvider) SayHello(ctx context.Context, name string) (string, error) {
	attachmentInfo := ctx.Value(constant.AttachmentKey).(map[string]interface{})
	fmt.Println(attachmentInfo)
	return "hello world " + name, nil
}

// @desc 测试注释2
// @used test
func (h *HelloProvider) SayAgain(id int, name string) (string, error) {
	fmt.Println("call SayAgain", id)
	return "hello world SayAgain" + name, nil
}
```
