// Code generated by SQLBoiler 4.11.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// Forum is an object representing the database table.
type Forum struct {
	ID int `boil:"id" json:"id" toml:"id" yaml:"id"`
	// The name given to the forum by the course administrator.
	Name      string    `boil:"name" json:"name" toml:"name" yaml:"name"`
	CreatedAt null.Time `boil:"created_at" json:"created_at,omitempty" toml:"created_at" yaml:"created_at,omitempty"`
	UpdatedAt null.Time `boil:"updated_at" json:"updated_at,omitempty" toml:"updated_at" yaml:"updated_at,omitempty"`
	DeletedAt null.Time `boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`

	R *forumR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L forumL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var ForumColumns = struct {
	ID        string
	Name      string
	CreatedAt string
	UpdatedAt string
	DeletedAt string
}{
	ID:        "id",
	Name:      "name",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
	DeletedAt: "deleted_at",
}

var ForumTableColumns = struct {
	ID        string
	Name      string
	CreatedAt string
	UpdatedAt string
	DeletedAt string
}{
	ID:        "forum.id",
	Name:      "forum.name",
	CreatedAt: "forum.created_at",
	UpdatedAt: "forum.updated_at",
	DeletedAt: "forum.deleted_at",
}

// Generated where

var ForumWhere = struct {
	ID        whereHelperint
	Name      whereHelperstring
	CreatedAt whereHelpernull_Time
	UpdatedAt whereHelpernull_Time
	DeletedAt whereHelpernull_Time
}{
	ID:        whereHelperint{field: "`forum`.`id`"},
	Name:      whereHelperstring{field: "`forum`.`name`"},
	CreatedAt: whereHelpernull_Time{field: "`forum`.`created_at`"},
	UpdatedAt: whereHelpernull_Time{field: "`forum`.`updated_at`"},
	DeletedAt: whereHelpernull_Time{field: "`forum`.`deleted_at`"},
}

// ForumRels is where relationship names are stored.
var ForumRels = struct {
	Courses      string
	ForumEntries string
}{
	Courses:      "Courses",
	ForumEntries: "ForumEntries",
}

// forumR is where relationships are stored.
type forumR struct {
	Courses      CourseSlice     `boil:"Courses" json:"Courses" toml:"Courses" yaml:"Courses"`
	ForumEntries ForumEntrySlice `boil:"ForumEntries" json:"ForumEntries" toml:"ForumEntries" yaml:"ForumEntries"`
}

// NewStruct creates a new relationship struct
func (*forumR) NewStruct() *forumR {
	return &forumR{}
}

func (r *forumR) GetCourses() CourseSlice {
	if r == nil {
		return nil
	}
	return r.Courses
}

func (r *forumR) GetForumEntries() ForumEntrySlice {
	if r == nil {
		return nil
	}
	return r.ForumEntries
}

// forumL is where Load methods for each relationship are stored.
type forumL struct{}

var (
	forumAllColumns            = []string{"id", "name", "created_at", "updated_at", "deleted_at"}
	forumColumnsWithoutDefault = []string{"name", "created_at", "updated_at", "deleted_at"}
	forumColumnsWithDefault    = []string{"id"}
	forumPrimaryKeyColumns     = []string{"id"}
	forumGeneratedColumns      = []string{}
)

type (
	// ForumSlice is an alias for a slice of pointers to Forum.
	// This should almost always be used instead of []Forum.
	ForumSlice []*Forum
	// ForumHook is the signature for custom Forum hook methods
	ForumHook func(context.Context, boil.ContextExecutor, *Forum) error

	forumQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	forumType                 = reflect.TypeOf(&Forum{})
	forumMapping              = queries.MakeStructMapping(forumType)
	forumPrimaryKeyMapping, _ = queries.BindMapping(forumType, forumMapping, forumPrimaryKeyColumns)
	forumInsertCacheMut       sync.RWMutex
	forumInsertCache          = make(map[string]insertCache)
	forumUpdateCacheMut       sync.RWMutex
	forumUpdateCache          = make(map[string]updateCache)
	forumUpsertCacheMut       sync.RWMutex
	forumUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var forumAfterSelectHooks []ForumHook

var forumBeforeInsertHooks []ForumHook
var forumAfterInsertHooks []ForumHook

var forumBeforeUpdateHooks []ForumHook
var forumAfterUpdateHooks []ForumHook

var forumBeforeDeleteHooks []ForumHook
var forumAfterDeleteHooks []ForumHook

var forumBeforeUpsertHooks []ForumHook
var forumAfterUpsertHooks []ForumHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *Forum) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range forumAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *Forum) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range forumBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *Forum) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range forumAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *Forum) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range forumBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *Forum) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range forumAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *Forum) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range forumBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *Forum) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range forumAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *Forum) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range forumBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *Forum) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range forumAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddForumHook registers your hook function for all future operations.
func AddForumHook(hookPoint boil.HookPoint, forumHook ForumHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		forumAfterSelectHooks = append(forumAfterSelectHooks, forumHook)
	case boil.BeforeInsertHook:
		forumBeforeInsertHooks = append(forumBeforeInsertHooks, forumHook)
	case boil.AfterInsertHook:
		forumAfterInsertHooks = append(forumAfterInsertHooks, forumHook)
	case boil.BeforeUpdateHook:
		forumBeforeUpdateHooks = append(forumBeforeUpdateHooks, forumHook)
	case boil.AfterUpdateHook:
		forumAfterUpdateHooks = append(forumAfterUpdateHooks, forumHook)
	case boil.BeforeDeleteHook:
		forumBeforeDeleteHooks = append(forumBeforeDeleteHooks, forumHook)
	case boil.AfterDeleteHook:
		forumAfterDeleteHooks = append(forumAfterDeleteHooks, forumHook)
	case boil.BeforeUpsertHook:
		forumBeforeUpsertHooks = append(forumBeforeUpsertHooks, forumHook)
	case boil.AfterUpsertHook:
		forumAfterUpsertHooks = append(forumAfterUpsertHooks, forumHook)
	}
}

// One returns a single forum record from the query.
func (q forumQuery) One(ctx context.Context, exec boil.ContextExecutor) (*Forum, error) {
	o := &Forum{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for forum")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all Forum records from the query.
func (q forumQuery) All(ctx context.Context, exec boil.ContextExecutor) (ForumSlice, error) {
	var o []*Forum

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to Forum slice")
	}

	if len(forumAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all Forum records in the query.
func (q forumQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count forum rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q forumQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if forum exists")
	}

	return count > 0, nil
}

// Courses retrieves all the course's Courses with an executor.
func (o *Forum) Courses(mods ...qm.QueryMod) courseQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("`course`.`forum_id`=?", o.ID),
	)

	return Courses(queryMods...)
}

// ForumEntries retrieves all the forum_entry's ForumEntries with an executor.
func (o *Forum) ForumEntries(mods ...qm.QueryMod) forumEntryQuery {
	var queryMods []qm.QueryMod
	if len(mods) != 0 {
		queryMods = append(queryMods, mods...)
	}

	queryMods = append(queryMods,
		qm.Where("`forum_entry`.`forum_id`=?", o.ID),
	)

	return ForumEntries(queryMods...)
}

// LoadCourses allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (forumL) LoadCourses(ctx context.Context, e boil.ContextExecutor, singular bool, maybeForum interface{}, mods queries.Applicator) error {
	var slice []*Forum
	var object *Forum

	if singular {
		object = maybeForum.(*Forum)
	} else {
		slice = *maybeForum.(*[]*Forum)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &forumR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &forumR{}
			}

			for _, a := range args {
				if a == obj.ID {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`course`),
		qm.WhereIn(`course.forum_id in ?`, args...),
		qmhelper.WhereIsNull(`course.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load course")
	}

	var resultSlice []*Course
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice course")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on course")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for course")
	}

	if len(courseAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.Courses = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &courseR{}
			}
			foreign.R.Forum = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.ForumID {
				local.R.Courses = append(local.R.Courses, foreign)
				if foreign.R == nil {
					foreign.R = &courseR{}
				}
				foreign.R.Forum = local
				break
			}
		}
	}

	return nil
}

// LoadForumEntries allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for a 1-M or N-M relationship.
func (forumL) LoadForumEntries(ctx context.Context, e boil.ContextExecutor, singular bool, maybeForum interface{}, mods queries.Applicator) error {
	var slice []*Forum
	var object *Forum

	if singular {
		object = maybeForum.(*Forum)
	} else {
		slice = *maybeForum.(*[]*Forum)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &forumR{}
		}
		args = append(args, object.ID)
	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &forumR{}
			}

			for _, a := range args {
				if a == obj.ID {
					continue Outer
				}
			}

			args = append(args, obj.ID)
		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`forum_entry`),
		qm.WhereIn(`forum_entry.forum_id in ?`, args...),
		qmhelper.WhereIsNull(`forum_entry.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load forum_entry")
	}

	var resultSlice []*ForumEntry
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice forum_entry")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results in eager load on forum_entry")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for forum_entry")
	}

	if len(forumEntryAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}
	if singular {
		object.R.ForumEntries = resultSlice
		for _, foreign := range resultSlice {
			if foreign.R == nil {
				foreign.R = &forumEntryR{}
			}
			foreign.R.Forum = object
		}
		return nil
	}

	for _, foreign := range resultSlice {
		for _, local := range slice {
			if local.ID == foreign.ForumID {
				local.R.ForumEntries = append(local.R.ForumEntries, foreign)
				if foreign.R == nil {
					foreign.R = &forumEntryR{}
				}
				foreign.R.Forum = local
				break
			}
		}
	}

	return nil
}

// AddCourses adds the given related objects to the existing relationships
// of the forum, optionally inserting them as new records.
// Appends related to o.R.Courses.
// Sets related.R.Forum appropriately.
func (o *Forum) AddCourses(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*Course) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.ForumID = o.ID
			if err = rel.Insert(ctx, exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE `course` SET %s WHERE %s",
				strmangle.SetParamNames("`", "`", 0, []string{"forum_id"}),
				strmangle.WhereClause("`", "`", 0, coursePrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.IsDebug(ctx) {
				writer := boil.DebugWriterFrom(ctx)
				fmt.Fprintln(writer, updateQuery)
				fmt.Fprintln(writer, values)
			}
			if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.ForumID = o.ID
		}
	}

	if o.R == nil {
		o.R = &forumR{
			Courses: related,
		}
	} else {
		o.R.Courses = append(o.R.Courses, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &courseR{
				Forum: o,
			}
		} else {
			rel.R.Forum = o
		}
	}
	return nil
}

// AddForumEntries adds the given related objects to the existing relationships
// of the forum, optionally inserting them as new records.
// Appends related to o.R.ForumEntries.
// Sets related.R.Forum appropriately.
func (o *Forum) AddForumEntries(ctx context.Context, exec boil.ContextExecutor, insert bool, related ...*ForumEntry) error {
	var err error
	for _, rel := range related {
		if insert {
			rel.ForumID = o.ID
			if err = rel.Insert(ctx, exec, boil.Infer()); err != nil {
				return errors.Wrap(err, "failed to insert into foreign table")
			}
		} else {
			updateQuery := fmt.Sprintf(
				"UPDATE `forum_entry` SET %s WHERE %s",
				strmangle.SetParamNames("`", "`", 0, []string{"forum_id"}),
				strmangle.WhereClause("`", "`", 0, forumEntryPrimaryKeyColumns),
			)
			values := []interface{}{o.ID, rel.ID}

			if boil.IsDebug(ctx) {
				writer := boil.DebugWriterFrom(ctx)
				fmt.Fprintln(writer, updateQuery)
				fmt.Fprintln(writer, values)
			}
			if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
				return errors.Wrap(err, "failed to update foreign table")
			}

			rel.ForumID = o.ID
		}
	}

	if o.R == nil {
		o.R = &forumR{
			ForumEntries: related,
		}
	} else {
		o.R.ForumEntries = append(o.R.ForumEntries, related...)
	}

	for _, rel := range related {
		if rel.R == nil {
			rel.R = &forumEntryR{
				Forum: o,
			}
		} else {
			rel.R.Forum = o
		}
	}
	return nil
}

// Forums retrieves all the records using an executor.
func Forums(mods ...qm.QueryMod) forumQuery {
	mods = append(mods, qm.From("`forum`"), qmhelper.WhereIsNull("`forum`.`deleted_at`"))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"`forum`.*"})
	}

	return forumQuery{q}
}

// FindForum retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindForum(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*Forum, error) {
	forumObj := &Forum{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from `forum` where `id`=? and `deleted_at` is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, forumObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from forum")
	}

	if err = forumObj.doAfterSelectHooks(ctx, exec); err != nil {
		return forumObj, err
	}

	return forumObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *Forum) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no forum provided for insertion")
	}

	var err error
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if queries.MustTime(o.CreatedAt).IsZero() {
			queries.SetScanner(&o.CreatedAt, currTime)
		}
		if queries.MustTime(o.UpdatedAt).IsZero() {
			queries.SetScanner(&o.UpdatedAt, currTime)
		}
	}

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(forumColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	forumInsertCacheMut.RLock()
	cache, cached := forumInsertCache[key]
	forumInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			forumAllColumns,
			forumColumnsWithDefault,
			forumColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(forumType, forumMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(forumType, forumMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO `forum` (`%s`) %%sVALUES (%s)%%s", strings.Join(wl, "`,`"), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO `forum` () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT `%s` FROM `forum` WHERE %s", strings.Join(returnColumns, "`,`"), strmangle.WhereClause("`", "`", 0, forumPrimaryKeyColumns))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	result, err := exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into forum")
	}

	var lastID int64
	var identifierCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	lastID, err = result.LastInsertId()
	if err != nil {
		return ErrSyncFail
	}

	o.ID = int(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == forumMapping["id"] {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.ID,
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, identifierCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "models: unable to populate default values for forum")
	}

CacheNoHooks:
	if !cached {
		forumInsertCacheMut.Lock()
		forumInsertCache[key] = cache
		forumInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the Forum.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *Forum) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		queries.SetScanner(&o.UpdatedAt, currTime)
	}

	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	forumUpdateCacheMut.RLock()
	cache, cached := forumUpdateCache[key]
	forumUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			forumAllColumns,
			forumPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update forum, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE `forum` SET %s WHERE %s",
			strmangle.SetParamNames("`", "`", 0, wl),
			strmangle.WhereClause("`", "`", 0, forumPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(forumType, forumMapping, append(wl, forumPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update forum row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for forum")
	}

	if !cached {
		forumUpdateCacheMut.Lock()
		forumUpdateCache[key] = cache
		forumUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q forumQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for forum")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for forum")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o ForumSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), forumPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE `forum` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, forumPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in forum slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all forum")
	}
	return rowsAff, nil
}

var mySQLForumUniqueColumns = []string{
	"id",
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *Forum) Upsert(ctx context.Context, exec boil.ContextExecutor, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no forum provided for upsert")
	}
	if !boil.TimestampsAreSkipped(ctx) {
		currTime := time.Now().In(boil.GetLocation())

		if queries.MustTime(o.CreatedAt).IsZero() {
			queries.SetScanner(&o.CreatedAt, currTime)
		}
		queries.SetScanner(&o.UpdatedAt, currTime)
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(forumColumnsWithDefault, o)
	nzUniques := queries.NonZeroDefaultSet(mySQLForumUniqueColumns, o)

	if len(nzUniques) == 0 {
		return errors.New("cannot upsert with a table that cannot conflict on a unique column")
	}

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzUniques {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	forumUpsertCacheMut.RLock()
	cache, cached := forumUpsertCache[key]
	forumUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			forumAllColumns,
			forumColumnsWithDefault,
			forumColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			forumAllColumns,
			forumPrimaryKeyColumns,
		)

		if !updateColumns.IsNone() && len(update) == 0 {
			return errors.New("models: unable to upsert forum, could not build update column list")
		}

		ret = strmangle.SetComplement(ret, nzUniques)
		cache.query = buildUpsertQueryMySQL(dialect, "`forum`", update, insert)
		cache.retQuery = fmt.Sprintf(
			"SELECT %s FROM `forum` WHERE %s",
			strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, ret), ","),
			strmangle.WhereClause("`", "`", 0, nzUniques),
		)

		cache.valueMapping, err = queries.BindMapping(forumType, forumMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(forumType, forumMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	result, err := exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "models: unable to upsert for forum")
	}

	var lastID int64
	var uniqueMap []uint64
	var nzUniqueCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	lastID, err = result.LastInsertId()
	if err != nil {
		return ErrSyncFail
	}

	o.ID = int(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == forumMapping["id"] {
		goto CacheNoHooks
	}

	uniqueMap, err = queries.BindMapping(forumType, forumMapping, nzUniques)
	if err != nil {
		return errors.Wrap(err, "models: unable to retrieve unique values for forum")
	}
	nzUniqueCols = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), uniqueMap)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, nzUniqueCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, nzUniqueCols...).Scan(returns...)
	if err != nil {
		return errors.Wrap(err, "models: unable to populate default values for forum")
	}

CacheNoHooks:
	if !cached {
		forumUpsertCacheMut.Lock()
		forumUpsertCache[key] = cache
		forumUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single Forum record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *Forum) Delete(ctx context.Context, exec boil.ContextExecutor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no Forum provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), forumPrimaryKeyMapping)
		sql = "DELETE FROM `forum` WHERE `id`=?"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE `forum` SET %s WHERE `id`=?",
			strmangle.SetParamNames("`", "`", 0, wl),
		)
		valueMapping, err := queries.BindMapping(forumType, forumMapping, append(wl, forumPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), valueMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from forum")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for forum")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q forumQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no forumQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from forum")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for forum")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o ForumSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(forumBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), forumPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM `forum` WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, forumPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), forumPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE `forum` SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, forumPrimaryKeyColumns, len(o)),
			strmangle.SetParamNames("`", "`", 0, wl),
		)
		args = append([]interface{}{currTime}, args...)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from forum slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for forum")
	}

	if len(forumAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *Forum) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindForum(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ForumSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := ForumSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), forumPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT `forum`.* FROM `forum` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, forumPrimaryKeyColumns, len(*o)) +
		"and `deleted_at` is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in ForumSlice")
	}

	*o = slice

	return nil
}

// ForumExists checks if the Forum row exists.
func ForumExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from `forum` where `id`=? and `deleted_at` is null limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if forum exists")
	}

	return exists, nil
}
