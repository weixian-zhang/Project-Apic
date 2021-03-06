package cmd

import (
	
	"io"
	"fmt"
	"net/http"
	"net/url"
	"bufio"
	"os"
	"strings"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)


const defaultPort string = "8080"
const defaultSwaggerPort string = "8090"
const defaultRestPath string = "/api/new"

// restCmd represents the rest command
var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "Creates REST API",
	Long: 
`
-apic host rest creates a Rest Api with a single path
-use with config file to create a Rest Api with many paths
-open multiple cmdlines to run multiple apic sessions to achieve multi-Apis and paths
Example default args: apic rest
Example 2 custom args: apic rest /user/new -p 8071 --swaggerport 8072 -q userid=110&tag=name -r {"userName":"John Nash"}`, //TODO desc flags
	Run: restCmdExecute,
}


func init() {

	rootCmd.AddCommand(restCmd)

	restCmd.PersistentFlags().StringP("config", "", "", "config file to host series of APIs")

	restCmd.PersistentFlags().StringP("swaggerport", "", "", "define Swagger docs serving port, default 8090")

	restCmd.PersistentFlags().StringP("querystr", "q", "", "query string: id=110&name='john dane'")
	
	restCmd.PersistentFlags().StringP("port", "p", "", "define listening port, default:8080")

	restCmd.PersistentFlags().StringP("header", "d", "", "i.e: content-type=application/json custom-key=customvalue")

	restCmd.PersistentFlags().StringP("cookie", "k", "", "i.e: cookie1=value1 cookie2=value2")
	
	restCmd.PersistentFlags().StringP("resp", "r", "", "define response (always json)")
}

func restCmdExecute(cmd *cobra.Command, args []string) {

	restApiContext := createApis(cmd)

	initApiListener(restApiContext)

	printAPIsInfo(restApiContext)

	pexit := make(chan bool, 1) //receive only
	pexit <- false

	go serveSwaggerDocs(pexit, restApiContext)

	reader := bufio.NewReader(os.Stdin)
	reader.ReadRune()

	pexit <- true //kills swagger.exe process

	fmt.Print("apic terminating...")
}

func createApis(cmd *cobra.Command) (RestApiContext) {

	restApiContext := newRestApiContext()

	port := getPort(cmd)
	restApiContext.Port = port
	
	sp,_ := cmd.Flags().GetString("swaggerport")
	if sp != "" {
		restApiContext.swaggerPort = sp
	}

	configPath, _ := cmd.Flags().GetString("config")
	restApiContext.configPath = configPath

	if configPath != "" {
		restApiContext.RestApis = createApisFromConfigFile(configPath)
	} else {
		restApiContext.RestApis = createApiFromCli(cmd)
	}

	return restApiContext
}

func createApiFromCli(cmd *cobra.Command) ([]RestApi) {

	apis := []RestApi{}

	api := newRestApi()
	api.Path = cmd.Flags().Arg(0)
	
	if api.Path == "" {
		api.Path = defaultRestPath
	} else {
		api.Path =  formatAPIPath(api.Path)
	}

	cmd.Flags().Visit(func(f *pflag.Flag) {

		switch f.Name {
			case "querystr":
				api.Querystring = formatQueryStr(f.Value.String())
			case "resp":
				api.Resp = f.Value.String()
			case "header":
				api.headerStr = strings.TrimSpace(f.Value.String())
				api.headers = newHeaderSlice(api.headerStr)
			case "cookie":
				api.cookieStr = strings.TrimSpace(f.Value.String())
				api.cookies = newCookieSlice(api.cookieStr)
		}
	})

	apis = append(apis, api)

	return apis
}

func createApisFromConfigFile(configPath string) ([]RestApi) {

	//TODO: log err
	fmt.Println(configPath)

	if configPath == "" {
		return nil
	}

	return nil
}

// func newApi(path string, qs string, resp string, headers string, cookies string) {

	
	

	
// }

func initApiListener(apiContext RestApiContext) {

	r := mux.NewRouter()

	for _, v := range apiContext.RestApis {

		resp := newResponse(v)

		createRestHandlers(r, v, resp)
	}

	go http.ListenAndServe(fmt.Sprintf(":%s", apiContext.Port), r)
}

func createRestHandlers(r *mux.Router, api RestApi, resp response ) { //cmd RestApi, cmdCon RestApiContext) {
	
	r.HandleFunc(api.Path, func(w http.ResponseWriter, r *http.Request){ handleResponse(w, r, resp) }).Methods("GET")
	
	r.HandleFunc(api.Path, func(w http.ResponseWriter, r *http.Request){ handleResponse(w, r, resp) }).Methods("POST")

	r.HandleFunc(api.Path, func(w http.ResponseWriter, r *http.Request){ handleResponse(w, r, resp) }).Methods("PUT")

	r.HandleFunc(api.Path, func(w http.ResponseWriter, r *http.Request){ handleResponse(w, r, resp) }).Methods("DELETE")
}

func handleResponse(w http.ResponseWriter, r *http.Request, resp response) { //} api RestApi, cmdCon RestApiContext) {

	for _, h := range resp.headers {
		w.Header().Add(h.key, h.value)
	}

	for _, c := range resp.cookies {
		cookie := http.Cookie{
			Name: c.name,
			Value: c.value,
		}

		http.SetCookie(w,&cookie)
	}

	newResp := fmt.Sprintf(`%v
%v`, r.Method, resp.resp)

	io.WriteString(w, newResp)

	printIngreReqInfo(r)
}

func getApiPath(args []string) string {
	if len(args) > 0 {
		return args[0]
	} else {
		return "/api/new"
	}
}

func formatAPIPath(path string) (string) {
	var newPath string
	newPath = strings.TrimSpace(path)

	if fc := newPath[0:1]; fc != "/" {
		newPath = "/" + newPath
	}

	return newPath
}

func formatQueryStr(qs string) (string) {

	fqs := url.QueryEscape(strings.TrimSpace(qs))
	if string(qs[0]) != "?" {
		return "?" + fqs
	} else {
		return fqs
	}
}

func getPort(cmd *cobra.Command) (string) {
	var port string = defaultPort
	p, _ := cmd.Flags().GetString("port")
	if p != "" {
		port = p
	}
	return strings.TrimSpace(port)
}


