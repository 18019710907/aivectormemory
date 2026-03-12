package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Issue struct {
	ID            int            `json:"id"`
	ProjectDir    string         `json:"project_dir"`
	IssueNumber   int            `json:"issue_number"`
	Date          string         `json:"date"`
	Title         string         `json:"title"`
	Status        string         `json:"status"`
	Content       string         `json:"content"`
	Tags          string         `json:"tags"`
	Description   string         `json:"description"`
	Investigation string         `json:"investigation"`
	RootCause     string         `json:"root_cause"`
	Solution      string         `json:"solution"`
	FilesChanged  string         `json:"files_changed"`
	TestResult    string         `json:"test_result"`
	Notes         string         `json:"notes"`
	FeatureID     string         `json:"feature_id"`
	ParentID      int            `json:"parent_id"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
	TaskProgress  map[string]int `json:"task_progress,omitempty"`
}

type IssueListResult struct {
	Issues     []Issue `json:"issues"`
	Total      int     `json:"total"`
	TotalAll   int     `json:"total_all"`
	TotalToday int     `json:"total_today"`
}

func (d *DB) GetIssues(projectDir, status, dateFrom, dateTo, keyword string, limit, offset int) (*IssueListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	today := time.Now().Format("2006-01-02")
	var totalAll, totalToday int
	d.QueryRow("SELECT COUNT(*) FROM (SELECT id FROM issues WHERE project_dir=? UNION ALL SELECT id FROM issues_archive WHERE project_dir=?)", projectDir, projectDir).Scan(&totalAll)
	d.QueryRow("SELECT COUNT(*) FROM (SELECT id FROM issues WHERE project_dir=? AND date=? UNION ALL SELECT id FROM issues_archive WHERE project_dir=? AND date=?)", projectDir, today, projectDir, today).Scan(&totalToday)

	if status == "archived" {
		return d.getArchivedIssues(projectDir, dateFrom, dateTo, keyword, limit, offset, totalAll, totalToday)
	}
	if status == "" || status == "all" {
		return d.getAllIssues(projectDir, dateFrom, dateTo, keyword, limit, offset, totalAll, totalToday)
	}

	where := []string{"project_dir=?"}
	args := []interface{}{projectDir}

	if status == "active" {
		where = append(where, "status IN ('pending','in_progress')")
	} else {
		where = append(where, "status=?")
		args = append(args, status)
	}
	dateWhere, dateArgs := dateRangeClause(dateFrom, dateTo)
	if dateWhere != "" {
		where = append(where, dateWhere)
		args = append(args, dateArgs...)
	}
	if keyword != "" {
		where = append(where, "(title LIKE ? OR content LIKE ?)")
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	d.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM issues WHERE %s", whereClause), args...).Scan(&total)

	args = append(args, limit, offset)
	rows, err := d.Query(fmt.Sprintf("SELECT * FROM issues WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?", whereClause), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := scanIssues(rows)
	d.attachTaskProgress(projectDir, issues)
	return &IssueListResult{Issues: issues, Total: total, TotalAll: totalAll, TotalToday: totalToday}, nil
}

func dateRangeClause(dateFrom, dateTo string) (string, []interface{}) {
	if dateFrom != "" && dateTo != "" {
		return "date >= ? AND date <= ?", []interface{}{dateFrom, dateTo}
	}
	if dateFrom != "" {
		return "date >= ?", []interface{}{dateFrom}
	}
	if dateTo != "" {
		return "date <= ?", []interface{}{dateTo}
	}
	return "", nil
}

func (d *DB) getArchivedIssues(projectDir, dateFrom, dateTo, keyword string, limit, offset int, totalAll, totalToday int) (*IssueListResult, error) {
	where := []string{"project_dir=?"}
	args := []interface{}{projectDir}
	dateWhere, dateArgs := dateRangeClause(dateFrom, dateTo)
	if dateWhere != "" {
		where = append(where, dateWhere)
		args = append(args, dateArgs...)
	}
	if keyword != "" {
		where = append(where, "(title LIKE ? OR content LIKE ?)")
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	whereClause := strings.Join(where, " AND ")

	var total int
	d.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM issues_archive WHERE %s", whereClause), args...).Scan(&total)

	args = append(args, limit, offset)
	rows, err := d.Query(fmt.Sprintf("SELECT id, project_dir, issue_number, date, title, content, description, investigation, root_cause, solution, files_changed, test_result, notes, feature_id, parent_id, status, created_at, archived_at as updated_at FROM issues_archive WHERE %s ORDER BY archived_at DESC LIMIT ? OFFSET ?", whereClause), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := scanIssues(rows)
	return &IssueListResult{Issues: issues, Total: total, TotalAll: totalAll, TotalToday: totalToday}, nil
}

func (d *DB) getAllIssues(projectDir, dateFrom, dateTo, keyword string, limit, offset int, totalAll, totalToday int) (*IssueListResult, error) {
	w1 := "WHERE project_dir=?"
	w2 := "WHERE project_dir=?"
	p1 := []interface{}{projectDir}
	p2 := []interface{}{projectDir}
	dateWhere, dateArgs := dateRangeClause(dateFrom, dateTo)
	if dateWhere != "" {
		w1 += " AND " + dateWhere
		w2 += " AND " + dateWhere
		p1 = append(p1, dateArgs...)
		p2 = append(p2, dateArgs...)
	}
	if keyword != "" {
		w1 += " AND (title LIKE ? OR content LIKE ?)"
		w2 += " AND (title LIKE ? OR content LIKE ?)"
		kw := "%" + keyword + "%"
		p1 = append(p1, kw, kw)
		p2 = append(p2, kw, kw)
	}

	allParams := append(append([]interface{}{}, p1...), p2...)

	var total int
	d.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM (SELECT id FROM issues %s UNION ALL SELECT id FROM issues_archive %s)", w1, w2), allParams...).Scan(&total)

	cols := "id, project_dir, issue_number, date, title, content, description, investigation, root_cause, solution, files_changed, test_result, notes, feature_id, parent_id, created_at"
	sql := fmt.Sprintf(
		"SELECT %s, status, updated_at FROM issues %s UNION ALL SELECT %s, 'archived' as status, archived_at as updated_at FROM issues_archive %s ORDER BY issue_number DESC LIMIT ? OFFSET ?",
		cols, w1, cols, w2)
	queryParams := append(allParams, limit, offset)
	rows, err := d.Query(sql, queryParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	issues := scanIssues(rows)
	d.attachTaskProgress(projectDir, issues)
	return &IssueListResult{Issues: issues, Total: total, TotalAll: totalAll, TotalToday: totalToday}, nil
}

func (d *DB) GetIssueDetail(id int, projectDir string) (*Issue, error) {
	row, err := d.Query("SELECT * FROM issues WHERE id=? AND project_dir=?", id, projectDir)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	issues := scanIssues(row)
	if len(issues) > 0 {
		return &issues[0], nil
	}
	// Fallback: check archived issues
	arow, err := d.Query("SELECT id, project_dir, issue_number, date, title, content, description, investigation, root_cause, solution, files_changed, test_result, notes, feature_id, parent_id, status, created_at, archived_at as updated_at FROM issues_archive WHERE id=? AND project_dir=?", id, projectDir)
	if err != nil {
		return nil, fmt.Errorf("issue not found: %d", id)
	}
	defer arow.Close()
	archived := scanIssues(arow)
	if len(archived) > 0 {
		return &archived[0], nil
	}
	return nil, fmt.Errorf("issue not found: %d", id)
}

func (d *DB) CreateIssue(projectDir, title, content, status string, tags []string, parentID int) (*Issue, bool, error) {
	if title == "" {
		return nil, false, fmt.Errorf("title required")
	}
	date := time.Now().Format("2006-01-02")
	now := time.Now().Format(time.RFC3339)
	if status == "" {
		status = "pending"
	}

	// Dedup check
	var dupCount int
	d.QueryRow("SELECT COUNT(*) FROM issues WHERE project_dir=? AND title=? AND status IN ('pending','in_progress')", projectDir, title).Scan(&dupCount)
	if dupCount > 0 {
		return nil, true, nil
	}

	// Get next issue_number
	var maxNum int
	d.QueryRow("SELECT COALESCE(MAX(issue_number),0) FROM issues WHERE project_dir=?", projectDir).Scan(&maxNum)

	tagsJSON := "[]"
	if len(tags) > 0 {
		if b, err := json.Marshal(tags); err == nil {
			tagsJSON = string(b)
		}
	}

	result, err := d.Exec(
		"INSERT INTO issues (project_dir, issue_number, date, title, status, content, tags, description, investigation, root_cause, solution, files_changed, test_result, notes, feature_id, parent_id, created_at, updated_at) VALUES (?,?,?,?,?,?,?,'','','','','[]','','','',?,?,?)",
		projectDir, maxNum+1, date, title, status, content, tagsJSON, parentID, now, now)
	if err != nil {
		return nil, false, err
	}

	id, _ := result.LastInsertId()
	issue, _ := d.GetIssueDetail(int(id), projectDir)
	return issue, false, nil
}

func (d *DB) UpdateIssue(id int, projectDir string, fields map[string]interface{}) (*Issue, error) {
	if len(fields) == 0 {
		return d.GetIssueDetail(id, projectDir)
	}

	sets := []string{}
	args := []interface{}{}
	for k, v := range fields {
		switch k {
		case "title", "status", "content", "description", "investigation",
			"root_cause", "solution", "test_result", "notes", "feature_id":
			sets = append(sets, k+"=?")
			args = append(args, v)
		case "tags", "files_changed":
			switch val := v.(type) {
			case string:
				sets = append(sets, k+"=?")
				args = append(args, val)
			default:
				b, _ := json.Marshal(val)
				sets = append(sets, k+"=?")
				args = append(args, string(b))
			}
		}
	}

	if len(sets) > 0 {
		sets = append(sets, "updated_at=?")
		args = append(args, time.Now().Format(time.RFC3339))
		args = append(args, id, projectDir)
		_, err := d.Exec(fmt.Sprintf("UPDATE issues SET %s WHERE id=? AND project_dir=?", strings.Join(sets, ",")), args...)
		if err != nil {
			return nil, err
		}
	}

	return d.GetIssueDetail(id, projectDir)
}

func (d *DB) ArchiveIssue(id int, projectDir string) error {
	row, err := d.Query("SELECT * FROM issues WHERE id=? AND project_dir=?", id, projectDir)
	if err != nil {
		return err
	}
	defer row.Close()
	issues := scanIssues(row)
	if len(issues) == 0 {
		return fmt.Errorf("issue not found: %d", id)
	}
	issue := issues[0]
	now := time.Now().Format(time.RFC3339)

	_, err = d.Exec(
		"INSERT INTO issues_archive (project_dir, issue_number, date, title, content, tags, description, investigation, root_cause, solution, files_changed, test_result, notes, feature_id, parent_id, status, original_issue_id, archived_at, created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		issue.ProjectDir, issue.IssueNumber, issue.Date, issue.Title, issue.Content,
		issue.Tags, issue.Description, issue.Investigation, issue.RootCause, issue.Solution,
		issue.FilesChanged, issue.TestResult, issue.Notes, issue.FeatureID,
		issue.ParentID, issue.Status, issue.ID, now, issue.CreatedAt)
	if err != nil {
		return err
	}

	_, err = d.Exec("DELETE FROM issues WHERE id=?", id)
	return err
}

func (d *DB) DeleteIssue(id int, projectDir string, archived bool) error {
	table := "issues"
	if archived {
		table = "issues_archive"
	}
	res, err := d.Exec(fmt.Sprintf("DELETE FROM %s WHERE id=? AND project_dir=?", table), id, projectDir)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("issue not found: %d", id)
	}
	return nil
}

func (d *DB) attachTaskProgress(projectDir string, issues []Issue) {
	fids := map[string]bool{}
	for _, i := range issues {
		if i.FeatureID != "" {
			fids[i.FeatureID] = true
		}
	}
	if len(fids) == 0 {
		return
	}

	progressMap := map[string]map[string]int{}
	for fid := range fids {
		var total, done int
		d.QueryRow("SELECT COUNT(*) FROM tasks WHERE project_dir=? AND feature_id=?", projectDir, fid).Scan(&total)
		d.QueryRow("SELECT COUNT(*) FROM tasks WHERE project_dir=? AND feature_id=? AND status='completed'", projectDir, fid).Scan(&done)
		progressMap[fid] = map[string]int{"total": total, "done": done}
	}

	for i := range issues {
		if p, ok := progressMap[issues[i].FeatureID]; ok {
			issues[i].TaskProgress = p
		}
	}
}

func scanIssues(rows *sql.Rows) []Issue {
	var issues []Issue
	cols, _ := rows.Columns()
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		rows.Scan(ptrs...)

		issue := Issue{}
		for i, col := range cols {
			s := toString(vals[i])
			switch col {
			case "id":
				if n, ok := vals[i].(int64); ok {
					issue.ID = int(n)
				}
			case "project_dir":
				issue.ProjectDir = s
			case "issue_number":
				if n, ok := vals[i].(int64); ok {
					issue.IssueNumber = int(n)
				}
			case "date":
				issue.Date = s
			case "title":
				issue.Title = s
			case "status":
				issue.Status = s
			case "content":
				issue.Content = s
			case "tags":
				if s == "" {
					s = "[]"
				}
				issue.Tags = s
			case "description":
				issue.Description = s
			case "investigation":
				issue.Investigation = s
			case "root_cause":
				issue.RootCause = s
			case "solution":
				issue.Solution = s
			case "files_changed":
				issue.FilesChanged = s
			case "test_result":
				issue.TestResult = s
			case "notes":
				issue.Notes = s
			case "feature_id":
				issue.FeatureID = s
			case "parent_id":
				if n, ok := vals[i].(int64); ok {
					issue.ParentID = int(n)
				}
			case "created_at":
				issue.CreatedAt = s
			case "updated_at":
				issue.UpdatedAt = s
			}
		}
		issues = append(issues, issue)
	}
	if issues == nil {
		issues = []Issue{}
	}
	return issues
}
