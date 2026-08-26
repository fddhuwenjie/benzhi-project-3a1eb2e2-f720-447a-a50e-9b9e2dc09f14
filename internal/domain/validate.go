package domain

import (
	"regexp"
	"strings"
	"time"
)

var materialCodePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{1,63}$`)

func ValidateCreate(in CreateInput) error {
	fields := map[string]string{}
	if strings.TrimSpace(in.Site) == "" {
		fields["site"] = "场所不能为空"
	}
	if strings.TrimSpace(in.Reason) == "" {
		fields["reason"] = "退役原因不能为空"
	}
	if strings.TrimSpace(in.OwnerID) == "" {
		fields["owner_id"] = "责任人不能为空"
	}
	if _, err := time.Parse("2006-01-02", in.PlannedDate); err != nil {
		fields["planned_date"] = "计划日期格式必须为 YYYY-MM-DD"
	}
	if len(in.Materials) == 0 {
		fields["materials"] = "至少登记一项材料"
	}
	seen := map[string]bool{}
	for i, material := range in.Materials {
		prefix := "materials[" + itoa(i) + "]."
		code := strings.TrimSpace(material.MaterialCode)
		if !materialCodePattern.MatchString(code) {
			fields[prefix+"material_code"] = "材料标识格式无效"
		}
		if seen[strings.ToLower(code)] {
			fields[prefix+"material_code"] = "材料标识必须唯一"
		}
		seen[strings.ToLower(code)] = true
		if strings.TrimSpace(material.DisplayName) == "" {
			fields[prefix+"display_name"] = "材料名称不能为空"
		}
		if material.DeclaredQuantity <= 0 {
			fields[prefix+"declared_quantity"] = "申报数量必须大于 0"
		}
		if strings.TrimSpace(material.Unit) == "" {
			fields[prefix+"unit"] = "单位不能为空"
		}
		if !oneOf(material.HazardClass, "general", "flammable", "corrosive", "toxic", "reactive", "biohazard") {
			fields[prefix+"hazard_class"] = "危险类别无效"
		}
		if !oneOf(material.PackageCondition, "intact", "damaged", "leaking") {
			fields[prefix+"package_condition"] = "包装状态无效"
		}
		if strings.TrimSpace(material.DisposalMethod) == "" {
			fields[prefix+"disposal_method"] = "处置方式不能为空"
		}
	}
	if len(fields) > 0 {
		return FieldError(fields)
	}
	return nil
}

func oneOf(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	result := ""
	for value > 0 {
		result = string(rune('0'+value%10)) + result
		value /= 10
	}
	return result
}

func uniquePeople(ids ...string) bool {
	seen := map[string]bool{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}
