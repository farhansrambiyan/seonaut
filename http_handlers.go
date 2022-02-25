package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/turk/go-sitemap"
	"golang.org/x/crypto/bcrypt"
)

const (
	inviteCode = "bATjGfQsRBeknDqD"
)

type ProjectView struct {
	Project Project
	Crawl   Crawl
}

type IssuesGroupView struct {
	IssuesGroups    map[string]IssueGroup
	Project         Project
	Crawl           Crawl
	MediaCount      CountList
	StatusCodeCount CountList
	MediaChart      Chart
	StatusChart     Chart
	Critical        int
	Alert           int
	Warning         int
}

type IssuesView struct {
	PageReports  []PageReport
	Cid          int
	Eid          string
	Project      Project
	CurrentPage  int
	NextPage     int
	PreviousPage int
	TotalPages   int
}

type ResourcesView struct {
	PageReport PageReport
	Cid        int
	Eid        string
	ErrorTypes []string
	InLinks    []PageReport
	Redirects  []PageReport
	Project    Project
	Tab        string
}

type Project struct {
	Id              int
	URL             string
	Host            string
	IgnoreRobotsTxt bool
	UseJS           bool
	Created         time.Time
}

func (app *App) serveHome(user *User, w http.ResponseWriter, r *http.Request) {
	var refresh bool
	var views []ProjectView
	projects := app.datastore.findProjectsByUser(user.Id)

	for _, p := range projects {
		c := app.datastore.getLastCrawl(&p)
		pv := ProjectView{
			Project: p,
			Crawl:   c,
		}
		views = append(views, pv)

		if c.IssuesEnd.Valid == false {
			refresh = true
		}
	}

	v := &PageView{
		Data: struct {
			Projects    []ProjectView
			MaxProjects int
		}{Projects: views, MaxProjects: user.getMaxAllowedProjects()},
		User:      *user,
		PageTitle: "PROJECTS_VIEW",
		Refresh:   refresh,
	}

	app.renderer.renderTemplate(w, "home", v)
}

func (app *App) serveProjectAdd(user *User, w http.ResponseWriter, r *http.Request) {
	projects := app.datastore.findProjectsByUser(user.Id)
	if len(projects) >= user.getMaxAllowedProjects() {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			log.Println(err)
		}

		url := r.FormValue("url")

		ignoreRobotsTxt, err := strconv.ParseBool(r.FormValue("ignore_robotstxt"))
		if err != nil {
			ignoreRobotsTxt = false
		}

		useJavascript, err := strconv.ParseBool(r.FormValue("use_javascript"))
		if err != nil {
			useJavascript = false
		}

		if user.Advanced == false {
			useJavascript = false
		}

		app.datastore.saveProject(url, ignoreRobotsTxt, useJavascript, user.Id)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &PageView{
		User:      *user,
		PageTitle: "ADD_PROJECT",
	}

	app.renderer.renderTemplate(w, "project_add", v)
}

func (app *App) serveCrawl(user *User, w http.ResponseWriter, r *http.Request) {
	qpid, ok := r.URL.Query()["pid"]
	if !ok || len(qpid) < 1 {
		log.Println("serveCrawl: pid parameter is missing")
		return
	}

	pid, err := strconv.Atoi(qpid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	p, err := app.datastore.findProjectById(pid, user.Id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	go func() {
		log.Printf("Crawling %s\n", p.URL)

		start := time.Now()
		cid := startCrawler(p, app.config.CrawlerAgent, user.Advanced, app.datastore, app.sanitizer)

		log.Printf("Done crawling %s in %s\n", p.URL, time.Since(start))
		log.Printf("Creating issues for %s and crawl id %d\n", p.URL, cid)

		rm := ReportManager{}

		rm.addReporter(app.datastore.Find30xPageReports, Error30x)
		rm.addReporter(app.datastore.Find40xPageReports, Error40x)
		rm.addReporter(app.datastore.Find50xPageReports, Error50x)
		rm.addReporter(app.datastore.FindPageReportsWithDuplicatedTitle, ErrorDuplicatedTitle)
		rm.addReporter(app.datastore.FindPageReportsWithDuplicatedTitle, ErrorDuplicatedDescription)
		rm.addReporter(app.datastore.FindPageReportsWithEmptyTitle, ErrorEmptyTitle)
		rm.addReporter(app.datastore.FindPageReportsWithShortTitle, ErrorShortTitle)
		rm.addReporter(app.datastore.FindPageReportsWithLongTitle, ErrorLongTitle)
		rm.addReporter(app.datastore.FindPageReportsWithEmptyDescription, ErrorEmptyDescription)
		rm.addReporter(app.datastore.FindPageReportsWithShortDescription, ErrorShortDescription)
		rm.addReporter(app.datastore.FindPageReportsWithLongDescription, ErrorLongDescription)
		rm.addReporter(app.datastore.FindPageReportsWithLittleContent, ErrorLittleContent)
		rm.addReporter(app.datastore.FindImagesWithNoAlt, ErrorImagesWithNoAlt)
		rm.addReporter(app.datastore.findRedirectChains, ErrorRedirectChain)
		rm.addReporter(app.datastore.FindPageReportsWithoutH1, ErrorNoH1)
		rm.addReporter(app.datastore.FindPageReportsWithNoLangAttr, ErrorNoLang)
		rm.addReporter(app.datastore.FindPageReportsWithHTTPLinks, ErrorHTTPLinks)
		rm.addReporter(app.datastore.FindMissingHrelangReturnLinks, ErrorHreflangsReturnLink)
		rm.addReporter(app.datastore.tooManyLinks, ErrorTooManyLinks)
		rm.addReporter(app.datastore.internalNoFollowLinks, ErrorInternalNoFollow)
		rm.addReporter(app.datastore.findExternalLinkWitoutNoFollow, ErrorExternalWithoutNoFollow)
		rm.addReporter(app.datastore.findCanonicalizedToNonCanonical, ErrorCanonicalizedToNonCanonical)
		rm.addReporter(app.datastore.findCanonicalizedToNonCanonical, ErrorRedirectLoop)
		rm.addReporter(app.datastore.findNotValidHeadingsOrder, ErrorNotValidHeadings)

		issues := rm.createIssues(cid)
		app.datastore.saveIssues(issues, cid)

		totalIssues := len(issues)

		app.datastore.saveEndIssues(cid, time.Now(), totalIssues)

		log.Printf("Done creating issues for %s...\n", p.URL)
		log.Printf("Deleting previous crawl data for %s\n", p.URL)
		app.datastore.DeletePreviousCrawl(p.Id)
		log.Printf("Deleted previous crawl done for %s\n", p.URL)
	}()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *App) serveIssues(user *User, w http.ResponseWriter, r *http.Request) {
	qcid, ok := r.URL.Query()["cid"]
	if !ok || len(qcid) < 1 {
		log.Println("serveIssues: cid parameter missing")
		return
	}

	cid, err := strconv.Atoi(qcid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	u, err := app.datastore.findCrawlUserId(cid)
	if err != nil || u.Id != user.Id {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	issueGroups := app.datastore.findIssues(cid)
	crawl := app.datastore.findCrawlById(cid)
	project, err := app.datastore.findProjectById(crawl.ProjectId, user.Id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	mediaCount := app.datastore.CountByMediaType(cid)
	mediaChart := NewChart(mediaCount)
	statusCount := app.datastore.CountByStatusCode(cid)
	statusChart := NewChart(statusCount)

	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	project.Host = parsedURL.Host

	var critical int
	var alert int
	var warning int

	for _, v := range issueGroups {
		switch v.Priority {
		case Critical:
			critical += v.Count
		case Alert:
			alert += v.Count
		case Warning:
			warning += v.Count
		}
	}

	ig := IssuesGroupView{
		IssuesGroups:    issueGroups,
		Crawl:           crawl,
		Project:         project,
		MediaCount:      mediaCount,
		MediaChart:      mediaChart,
		StatusChart:     statusChart,
		StatusCodeCount: statusCount,
		Critical:        critical,
		Alert:           alert,
		Warning:         warning,
	}

	v := &PageView{
		Data:      ig,
		User:      *user,
		PageTitle: "ISSUES_VIEW",
	}

	app.renderer.renderTemplate(w, "issues", v)
}

func (app *App) serveIssuesView(user *User, w http.ResponseWriter, r *http.Request) {
	qeid, ok := r.URL.Query()["eid"]
	if !ok || len(qeid) < 1 {
		log.Println("serveIssuesView: eid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	eid := qeid[0]

	qcid, ok := r.URL.Query()["cid"]
	if !ok || len(qcid) < 1 {
		log.Println("serveIssuesView: cid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	cid, err := strconv.Atoi(qcid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	totalPages := app.datastore.getNumberOfPagesForIssues(cid, eid)

	p := r.URL.Query()["p"]
	page := 1
	if len(p) > 0 {
		page, err = strconv.Atoi(p[0])
		if err != nil {
			log.Println(err)
			page = 1
		}

		if page < 1 || page > totalPages {
			http.Redirect(w, r, "/", http.StatusSeeOther)

			return
		}
	}

	nextPage := 0
	previousPage := 0

	if page < totalPages {
		nextPage = page + 1
	}

	if page > 1 {
		previousPage = page - 1
	}

	u, err := app.datastore.findCrawlUserId(cid)
	if err != nil || u.Id != user.Id {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	crawl := app.datastore.findCrawlById(cid)
	project, err := app.datastore.findProjectById(crawl.ProjectId, user.Id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	project.Host = parsedURL.Host

	issues := app.datastore.findPageReportIssues(cid, page-1, eid)

	view := IssuesView{
		Cid:          cid,
		Eid:          eid,
		PageReports:  issues,
		Project:      project,
		CurrentPage:  page,
		NextPage:     nextPage,
		PreviousPage: previousPage,
		TotalPages:   totalPages,
	}

	v := &PageView{
		Data:      view,
		User:      *user,
		PageTitle: "ISSUES_DETAIL",
	}

	app.renderer.renderTemplate(w, "issues_view", v)
}

func (app *App) serveResourcesView(user *User, w http.ResponseWriter, r *http.Request) {
	qrid, ok := r.URL.Query()["rid"]
	if !ok || len(qrid) < 1 {
		log.Println("serveResourcesView: rid paramenter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	rid, err := strconv.Atoi(qrid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	qcid, ok := r.URL.Query()["cid"]
	if !ok || len(qcid) < 1 {
		log.Println("serveResourcesView: cid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	cid, err := strconv.Atoi(qcid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	tabs := r.URL.Query()["t"]
	var tab string
	if len(tabs) == 0 {
		tab = "details"
	} else {
		tab = tabs[0]
	}

	qeid, ok := r.URL.Query()["eid"]
	if !ok || len(qeid) < 1 {
		log.Println("serveResourcesView: eid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	eid := qeid[0]

	u, err := app.datastore.findCrawlUserId(cid)
	if err != nil || u.Id != user.Id {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	crawl := app.datastore.findCrawlById(cid)
	project, err := app.datastore.findProjectById(crawl.ProjectId, user.Id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	project.Host = parsedURL.Host

	pageReport := app.datastore.FindPageReportById(rid)
	errorTypes := app.datastore.findErrorTypesByPage(rid, cid)

	var inLinks []PageReport
	if tab == "inlinks" {
		inLinks = app.datastore.FindInLinks(pageReport.URL, cid)
	}

	var redirects []PageReport
	if tab == "redirections" {
		redirects = app.datastore.FindPageReportsRedirectingToURL(pageReport.URL, cid)
	}

	rv := ResourcesView{
		PageReport: pageReport,
		Project:    project,
		Cid:        cid,
		Eid:        eid,
		ErrorTypes: errorTypes,
		InLinks:    inLinks,
		Redirects:  redirects,
		Tab:        tab,
	}

	v := &PageView{
		Data:      rv,
		User:      *user,
		PageTitle: "RESOURCES_VIEW",
	}

	app.renderer.renderTemplate(w, "resources", v)
}

func (app *App) serveSignup(w http.ResponseWriter, r *http.Request) {
	var invite bool

	inviteQ := r.URL.Query()["invite"]
	if len(inviteQ) == 0 {
		invite = false
	} else {
		if inviteQ[0] == inviteCode {
			invite = true
		}
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/signup", http.StatusSeeOther)

			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		exists := app.datastore.emailExists(email)
		if exists || password == "" {
			app.renderer.renderTemplate(w, "signup", &PageView{
				PageTitle: "SIGNUP_VIEW",
				Data: struct {
					Invite, Error bool
					Email         string
				}{Invite: invite, Error: true, Email: email},
			})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/signup", http.StatusSeeOther)

			return
		}

		app.datastore.userSignup(email, string(hashedPassword))

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	app.renderer.renderTemplate(w, "signup", &PageView{
		PageTitle: "SIGNUP_VIEW",
		Data: struct {
			Invite, Error bool
			Email         string
		}{Invite: invite, Error: false, Email: ""},
	})
}

func (app *App) serveSignin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)

			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		u := app.datastore.findUserByEmail(email)
		if u.Id == 0 {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		session, _ := app.cookie.Get(r, "SESSION_ID")
		session.Values["authenticated"] = true
		session.Values["uid"] = u.Id
		session.Save(r, w)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	v := &PageView{
		PageTitle: "SIGNIN_VIEW",
	}

	app.renderer.renderTemplate(w, "signin", v)
}

func (app *App) serveDownloadCSV(user *User, w http.ResponseWriter, r *http.Request) {
	qcid, ok := r.URL.Query()["cid"]
	if !ok || len(qcid) < 1 {
		log.Println("serveDownloadCSV: cid parameter missing")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	cid, err := strconv.Atoi(qcid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	u, err := app.datastore.findCrawlUserId(cid)
	if err != nil || u.Id != user.Id {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	crawl := app.datastore.findCrawlById(cid)

	project, err := app.datastore.findProjectById(crawl.ProjectId, user.Id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	var pageReports []PageReport

	eid := r.URL.Query()["eid"]
	fileName := parsedURL.Host + " crawl " + time.Now().Format("2-15-2006")

	if len(eid) > 0 && eid[0] != "" {
		fileName = fileName + "-" + eid[0]
		pageReports = app.datastore.FindAllPageReportsByCrawlIdAndErrorType(cid, eid[0])
	} else {
		pageReports = app.datastore.FindAllPageReportsByCrawlId(cid)
	}

	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", fileName))

	cw := newCSVWriter(w)
	for _, p := range pageReports {
		cw.write(p)
	}
}

func (app *App) serveSitemap(user *User, w http.ResponseWriter, r *http.Request) {
	qcid, ok := r.URL.Query()["cid"]
	if !ok || len(qcid) < 1 {
		log.Println("serveSitemap: cid parameter missings")
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	cid, err := strconv.Atoi(qcid[0])
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	u, err := app.datastore.findCrawlUserId(cid)
	if err != nil || u.Id != user.Id {
		http.Redirect(w, r, "/", http.StatusSeeOther)

		return
	}

	crawl := app.datastore.findCrawlById(cid)
	project, err := app.datastore.findProjectById(crawl.ProjectId, user.Id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	w.Header().Add(
		"Content-Disposition",
		fmt.Sprint("attachment; filename=\""+parsedURL.Host+" "+time.Now().Format("2-15-2006")+" sitemap.xml\""))

	s := sitemap.NewSitemap(w, true)
	p := app.datastore.findSitemapPageReports(cid)
	for _, v := range p {
		s.Add(v.URL, "")
	}

	s.Write()
}

func (app *App) serveSignout(user *User, w http.ResponseWriter, r *http.Request) {
	session, _ := app.cookie.Get(r, "SESSION_ID")
	session.Values["authenticated"] = false
	session.Values["uid"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *App) requireAuth(f func(user *User, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := app.cookie.Get(r, "SESSION_ID")
		var authenticated interface{} = session.Values["authenticated"]
		if authenticated != nil {
			isAuthenticated := session.Values["authenticated"].(bool)
			if isAuthenticated {
				session, _ := app.cookie.Get(r, "SESSION_ID")
				uid := session.Values["uid"].(int)
				user := app.datastore.findUserById(uid)
				f(user, w, r)

				return
			}
		}

		http.Redirect(w, r, "/signin", http.StatusSeeOther)
	}
}
