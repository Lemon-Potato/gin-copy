package gin

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"path"
)

type (
	HandlerFunc func(*Content)

	ErrorMsg struct {
		Message string `json:"msg"`
		Meta interface{} `json:"meta"`
	}

	ResponseWriter interface {
		http.ResponseWriter
		Status() int
		Written() bool
	}

	responseWriter struct {
		http.ResponseWriter
		status int
	}

	Content struct {
		Req *http.Request
		Writer ResponseWriter
		Error []ErrorMsg
		Params httprouter.Params
		handlers []HandlerFunc
		engine *Engine
		index int8
	}

	RouterGroup struct {
		Handlers []HandlerFunc
		prefix string
		parent *RouterGroup
		engine *Engine
	}

	Engine struct {
		*RouterGroup
		router *httprouter.Router
	}
)
// 写入 Header
func (rw *responseWriter) WriteHeader(s int) {
	rw.ResponseWriter.WriteHeader(s)
	rw.status = s
}
// 书写返回数据
func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
// 获取返回状态
func (rw *responseWriter) Status() int {
	return rw.status
}
// 是否完成返回值书写
func (rw *responseWriter) Written() bool {
	return rw.status != 0
}
// 创建引擎
func New() *Engine {
	engine := &Engine{}
	engine.RouterGroup = &RouterGroup{nil, "/", nil, engine}
	engine.router = httprouter.New()
	return engine
}
// 默认引擎
func Default() *Engine {
	engine := New()
	return engine
}
// 获取父路由组可能有的 请求处理（Handle）
func (group *RouterGroup) allHandlers(handlers []HandlerFunc) []HandlerFunc {
	local := append(group.Handlers, handlers...)
	if group.parent != nil {
		return group.parent.allHandlers(local)
	} else {
		return local
	}
}
// 创建响应 struct
func (group *RouterGroup) createContext(w http.ResponseWriter, req *http.Request, params httprouter.Params, handlers []HandlerFunc) *Content {
	return &Content{
		Req: req,
		Writer: &responseWriter{w, 0},
		index: -1,
		engine: group.engine,
		Params: params,
		handlers: handlers,
	}
}

// for 循环处理每一层的 Handle
func (c *Content) Next() {
	c.index++
	s := int8(len(c.handlers))
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}
// 最基本的响应添加
func (group *RouterGroup) Handle(method, p string, handlers []HandlerFunc) {
	p = path.Join(group.prefix, p)
	handlers = group.allHandlers(handlers)
	group.engine.router.Handle(method, p, func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		group.createContext(w, req, params, handlers).Next()
	})
}
// Handle 封装
func (group *RouterGroup) GET(path string, handlers ...HandlerFunc) {
	group.Handle("GET", path, handlers)
}
// http 响应接口实现
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	engine.router.ServeHTTP(w, req)
}
// 服务启动！
func (engine *Engine) Run(addr string) {
	err := http.ListenAndServe(addr, engine)
	if err != nil {
		fmt.Println(err)
	}
}
// 返回数据格式封装
func (c *Content) String(code int, msg string) {
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.Writer.WriteHeader(code)
	c.Writer.Write([]byte(msg))
}