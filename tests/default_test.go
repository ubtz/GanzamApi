package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	conf "GanzamApi/conf"
	"GanzamApi/models"
	"GanzamApi/repositories"
	"GanzamApi/services"
	"github.com/astaxie/beego/logs"

	"GanzamApi/controllers"
	_ "GanzamApi/routers"

	beego "github.com/astaxie/beego"
	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	_, file, _, _ := runtime.Caller(0)
	apppath, _ := filepath.Abs(filepath.Dir(filepath.Join(file, ".." + string(filepath.Separator))))
	beego.TestBeegoInit(apppath)
	controllers.SetAuthService(services.NewAuthService(repositories.NewMemoryUserStore()))
}

// TestBeego is a sample to run an endpoint test
func TestBeego(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	logs.Trace("testing", "TestBeego", "Code[%d]\n%s", w.Code, w.Body.String())

	Convey("Subject: Test Station Endpoint\n", t, func() {
	        Convey("Status Code Should Be 200", func() {
	                So(w.Code, ShouldEqual, 200)
	        })
	        Convey("The Result Should Not Be Empty", func() {
	                So(w.Body.Len(), ShouldBeGreaterThan, 0)
	        })
	})
}

func TestVersionEndpoint(t *testing.T) {
	_ = os.Setenv("APP_ENV", conf.EnvTest)
	_ = os.Setenv("TEST_API_URL", "https://test.ganzam.local")

	r, _ := http.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)

	Convey("Subject: Version Endpoint\n", t, func() {
		Convey("Status Code Should Be 200", func() {
			So(w.Code, ShouldEqual, 200)
		})
		Convey("Response Should Return Current Version", func() {
			So(err, ShouldBeNil)
			So(body["version"], ShouldEqual, controllers.CurrentVersion)
			So(body["environment"], ShouldEqual, conf.EnvTest)
			So(body["target_url"], ShouldEqual, "https://test.ganzam.local")
		})
	})
}

func TestRegisterEndpoint(t *testing.T) {
	controllers.SetAuthService(services.NewAuthService(repositories.NewMemoryUserStore()))

	body := []byte(`{"phone":"99112233","email":"user@test.com","password":"secret123","first_name":"Test","last_name":"User"}`)
	r, _ := http.NewRequest("POST", "/post/register", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)

	Convey("Subject: Register Endpoint\n", t, func() {
		Convey("Status Code Should Be 200", func() {
			So(w.Code, ShouldEqual, 200)
		})
		Convey("Register Should Return Token And User", func() {
			So(err, ShouldBeNil)
			So(response["token"], ShouldNotBeBlank)
			user, ok := response["user"].(map[string]interface{})
			So(ok, ShouldBeTrue)
			So(user["phone"], ShouldEqual, "99112233")
			So(user["role"], ShouldEqual, "customer")
		})
	})
}

func TestLoginEndpoint(t *testing.T) {
	store := repositories.NewMemoryUserStore()
	authService := services.NewAuthService(store)
	controllers.SetAuthService(authService)

	_, _, err := authService.Register(context.Background(), models.RegisterRequest{
		Phone:    "88110022",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("failed to seed login user: %v", err)
	}

	body := []byte(`{"phone":"88110022","password":"secret123"}`)
	r, _ := http.NewRequest("POST", "/post/login", bytes.NewBuffer(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	var response map[string]interface{}
	parseErr := json.Unmarshal(w.Body.Bytes(), &response)

	Convey("Subject: Login Endpoint\n", t, func() {
		Convey("Status Code Should Be 200", func() {
			So(w.Code, ShouldEqual, 200)
		})
		Convey("Login Should Return Token And User", func() {
			So(parseErr, ShouldBeNil)
			So(response["token"], ShouldNotBeBlank)
			user, ok := response["user"].(map[string]interface{})
			So(ok, ShouldBeTrue)
			So(user["phone"], ShouldEqual, "88110022")
		})
	})
}
