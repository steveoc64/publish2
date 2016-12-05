package publish2

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qor/qor/utils"
)

const (
	ModeOff             = "off"
	VersionMode         = "publish:version:mode"
	VersionNameMode     = "publish:version:name"
	VersionMultipleMode = "multiple"

	ScheduleMode   = "publish:schedule:mode"
	ScheduledTime  = "publish:schedule:current"
	ScheduledStart = "publish:schedule:start"
	ScheduledEnd   = "publish:schedule:end"

	VisibleMode = "publish:visible:mode"
)

func IsSchedulableModel(model interface{}) (ok bool) {
	if model != nil {
		_, ok = reflect.New(utils.ModelType(model)).Interface().(ScheduledInterface)
	}
	return
}

func IsVersionableModel(model interface{}) (ok bool) {
	if model != nil {
		_, ok = reflect.New(utils.ModelType(model)).Interface().(VersionableInterface)
	}
	return
}

func IsShareableVersionModel(model interface{}) (ok bool) {
	if model != nil {
		_, ok = reflect.New(utils.ModelType(model)).Interface().(ShareableVersionInterface)
	}
	return
}

func IsPublishReadyableModel(model interface{}) (ok bool) {
	if model != nil {
		_, ok = reflect.New(utils.ModelType(model)).Interface().(PublishReadyInterface)
	}
	return
}

func RegisterCallbacks(db *gorm.DB) {
	db.Callback().Query().Before("gorm:query").Register("publish:query", queryCallback)
	db.Callback().Query().After("gorm:preload").Register("publish:fix_preload", fixPreloadCallback)
	db.Callback().RowQuery().Before("gorm:query").Register("publish:query", queryCallback)

	db.Callback().Create().Before("gorm:begin_transaction").Register("publish:versions", createCallback)
	db.Callback().Update().Before("gorm:begin_transaction").Register("publish:versions", updateCallback)

	db.Callback().Delete().Before("gorm:begin_transaction").Register("publish:versions", deleteCallback)
}

func queryCallback(scope *gorm.Scope) {
	var (
		isSchedulable      = IsSchedulableModel(scope.Value)
		isVersionable      = IsVersionableModel(scope.Value)
		isShareableVersion = IsShareableVersionModel(scope.Value)
		isPublishReadyable = IsPublishReadyableModel(scope.Value)
		conditions         []string
		conditionValues    []interface{}
	)

	if isSchedulable {
		var scheduledStartTime, scheduledEndTime, scheduledCurrentTime *time.Time
		var mode, _ = scope.DB().Get(ScheduleMode)

		if v, ok := scope.Get(ScheduledStart); ok {
			if t, ok := v.(*time.Time); ok {
				scheduledStartTime = t
			} else if t, ok := v.(time.Time); ok {
				scheduledStartTime = &t
			}

			if scheduledStartTime != nil {
				conditions = append(conditions, "(scheduled_end_at IS NULL OR scheduled_end_at >= ?)")
				conditionValues = append(conditionValues, scheduledStartTime)
			}
		}

		if v, ok := scope.Get(ScheduledEnd); ok {
			if t, ok := v.(*time.Time); ok {
				scheduledEndTime = t
			} else if t, ok := v.(time.Time); ok {
				scheduledEndTime = &t
			}

			if scheduledEndTime != nil {
				conditions = append(conditions, "(scheduled_start_at IS NULL OR scheduled_start_at <= ?)")
				conditionValues = append(conditionValues, scheduledEndTime)
			}
		}

		if len(conditions) == 0 && mode != ModeOff {
			if v, ok := scope.Get(ScheduledTime); ok {
				if t, ok := v.(*time.Time); ok {
					scheduledCurrentTime = t
				} else if t, ok := v.(time.Time); ok {
					scheduledCurrentTime = &t
				}
			}

			if scheduledCurrentTime == nil {
				now := time.Now()
				scheduledCurrentTime = &now
			}

			conditions = append(conditions, "(scheduled_start_at IS NULL OR scheduled_start_at <= ?) AND (scheduled_end_at IS NULL OR scheduled_end_at >= ?)")
			conditionValues = append(conditionValues, scheduledCurrentTime, scheduledCurrentTime)
		}
	}

	if isPublishReadyable {
		switch mode, _ := scope.DB().Get(VisibleMode); mode {
		case ModeOff:
		default:
			conditions = append(conditions, "publish_ready = ?")
			conditionValues = append(conditionValues, true)
		}
	}

	if isVersionable {
		switch mode, _ := scope.DB().Get(VersionMode); mode {
		case VersionMultipleMode:
			scope.Search.Where(strings.Join(conditions, " AND "), conditionValues...)
		default:
			if versionName, ok := scope.DB().Get(VersionNameMode); ok && versionName != "" {
				scope.Search.Where("version_name = ?", versionName)
			} else {
				var sql string
				var primaryKeys []string

				if scope.HasColumn("DeletedAt") {
					conditions = append(conditions, "deleted_at IS NULL")
				}

				for _, primaryField := range scope.PrimaryFields() {
					if primaryField.DBName != "version_name" {
						primaryKeys = append(primaryKeys, scope.Quote(primaryField.DBName))
					}
				}

				primaryKeyCondition := strings.Join(primaryKeys, ",")
				if len(conditions) == 0 {
					sql = fmt.Sprintf("(%v, version_priority) IN (SELECT %v, MAX(version_priority) FROM %v GROUP BY %v)", primaryKeyCondition, primaryKeyCondition, scope.QuotedTableName(), primaryKeyCondition)
				} else {
					sql = fmt.Sprintf("(%v, version_priority) IN (SELECT %v, MAX(version_priority) FROM %v WHERE %v GROUP BY %v)", primaryKeyCondition, primaryKeyCondition, scope.QuotedTableName(), strings.Join(conditions, " AND "), primaryKeyCondition)
				}

				scope.Search.Where(sql, conditionValues...)
			}
		}

		scope.Search.Order("version_priority DESC")
	} else {
		if isShareableVersion {
			if versionName, ok := scope.DB().Get(VersionNameMode); ok && versionName != "" {
				var primaryKeys []string
				for _, primaryField := range scope.PrimaryFields() {
					if primaryField.DBName != "version_name" {
						primaryKeys = append(primaryKeys, scope.Quote(primaryField.DBName))
					}
				}

				primaryKeyCondition := strings.Join(primaryKeys, ",")

				scope.Search.Where(
					fmt.Sprintf("version_name = ? OR (version_name = ? AND (%v) NOT IN (SELECT %v FROM %v WHERE version_name = ?))", primaryKeyCondition, primaryKeyCondition, scope.QuotedTableName()),
					versionName, "", versionName,
				)
			}
		}

		scope.Search.Where(strings.Join(conditions, " AND "), conditionValues...)
	}
}

func fixPreloadCallback(scope *gorm.Scope) {
	filterFilterValuesWithVersion := func(gormField *gorm.Field, versionName string) {
		indirectFieldValue := reflect.Indirect(gormField.Field)
		switch indirectFieldValue.Kind() {
		case reflect.Slice:
			results := map[string]reflect.Value{}

			for i := 0; i < indirectFieldValue.Len(); i++ {
				if shareableVersion, ok := indirectFieldValue.Index(i).Addr().Interface().(ShareableVersionInterface); ok {
					fieldPrimaryValue := fmt.Sprint(scope.New(shareableVersion).PrimaryKeyValue())
					if _, ok := results[fieldPrimaryValue]; !ok || shareableVersion.GetSharedVersionName() == versionName {
						if shareableVersion.GetSharedVersionName() == versionName || shareableVersion.GetSharedVersionName() == "" {
							results[fieldPrimaryValue] = indirectFieldValue.Index(i)
						}
					}
				}
			}

			fieldResults := reflect.New(indirectFieldValue.Type()).Elem()
			for _, v := range results {
				fieldResults = reflect.Append(fieldResults, v)
			}
			gormField.Set(fieldResults)
		case reflect.Struct:
			if shareableVersion, ok := indirectFieldValue.Interface().(ShareableVersionInterface); ok {
				if shareableVersion.GetSharedVersionName() != "" && shareableVersion.GetSharedVersionName() != versionName {
					gormField.Set(reflect.New(indirectFieldValue.Type()))
				}
			}
		}
	}

	fixSharedVersionRecords := func(value interface{}, fieldName string) {
		reflectValue := reflect.Indirect(reflect.ValueOf(value))

		switch reflectValue.Kind() {
		case reflect.Slice:
			for i := 0; i < reflectValue.Len(); i++ {
				v := reflectValue.Index(i)
				if versionable, ok := v.Interface().(VersionableInterface); ok {
					if fieldValue, ok := scope.New(v.Interface()).FieldByName(fieldName); ok {
						filterFilterValuesWithVersion(fieldValue, versionable.GetVersionName())
					}
				}
			}
		case reflect.Struct:
			if versionable, ok := value.(VersionableInterface); ok {
				if fieldValue, ok := scope.New(value).FieldByName(fieldName); ok {
					filterFilterValuesWithVersion(fieldValue, versionable.GetVersionName())
				}
			}
		}
	}

	if IsVersionableModel(scope.Value) {
		for _, field := range scope.Fields() {
			if IsShareableVersionModel(reflect.New(field.Struct.Type).Interface()) {
				fixSharedVersionRecords(scope.Value, field.Name)
			}
		}
	}
}

func createCallback(scope *gorm.Scope) {
	if IsVersionableModel(scope.Value) {
		if field, ok := scope.FieldByName("VersionName"); ok {
			if field.IsBlank {
				field.Set(DefaultVersionName)
			}
		}

		updateVersionPriority(scope)
	}

	if IsShareableVersionModel(scope.Value) {
		if field, ok := scope.FieldByName("VersionName"); ok {
			field.IsBlank = false
		}
	}
}

func updateCallback(scope *gorm.Scope) {
	if IsVersionableModel(scope.Value) {
		updateVersionPriority(scope)
	}
}

func deleteCallback(scope *gorm.Scope) {
	if versionName, ok := scope.DB().Get(VersionNameMode); ok && versionName != "" {
		if IsVersionableModel(scope.Value) || IsShareableVersionModel(scope.Value) {
			scope.Search.Where("version_name = ?", versionName)
		}
	}
}

func updateVersionPriority(scope *gorm.Scope) {
	if field, ok := scope.FieldByName("VersionPriority"); ok {
		var scheduledTime *time.Time
		if scheduled, ok := scope.Value.(ScheduledInterface); ok {
			scheduledTime = scheduled.GetScheduledStartAt()
		}
		if scheduledTime == nil {
			unix := time.Unix(0, 0)
			scheduledTime = &unix
		}

		priority := fmt.Sprintf("%v_%v_%v", scope.PrimaryKeyValue(), scheduledTime.UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339Nano))
		field.Set(priority)
	}
}
