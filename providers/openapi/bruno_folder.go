package openapi

import (
	"mokapi/export/bruno"
	"path"
	"strings"
)

type brunoFolderBuilder struct {
	name        string
	summary     string
	description string

	ops    []*Operation
	order  []string
	items  map[string]*brunoFolderBuilder
	parent *brunoFolderBuilder
}

func (f *brunoFolderBuilder) insertTag(path []*Tag, op *Operation) {
	t := path[0]
	if len(path) == 1 {
		child := f.child(t.Name, t.Description, t.Summary)
		child.ops = append(child.ops, op)
		return
	}
	f.child(t.Name, t.Description, t.Summary).insertTag(path[1:], op)
}

func (f *brunoFolderBuilder) child(name, summary, description string) *brunoFolderBuilder {
	if _, ok := f.items[name]; !ok {
		f.items[name] = &brunoFolderBuilder{
			name:        name,
			summary:     summary,
			description: description,
			items:       map[string]*brunoFolderBuilder{},
			parent:      f,
		}
		f.order = append(f.order, name)
	}
	return f.items[name]
}

func (f *brunoFolderBuilder) insertPath(path []string, op *Operation) {
	if len(path) == 1 {
		f.ops = append(f.ops, op)
		return
	}
	if i, ok := f.items[path[0]]; ok {
		if i.name != path[0] {
			// split folder path into two folders
			newChildName := strings.TrimPrefix(i.name, path[0]+"/")
			child := &brunoFolderBuilder{name: newChildName, order: i.order, items: i.items, ops: i.ops, parent: i}
			i.name = path[0]
			i.order = []string{newChildName}
			i.items = map[string]*brunoFolderBuilder{newChildName: child}
			i.ops = nil
		}

		i.insertPath(path[1:], op)
		return
	}
	var name string
	i := 0
	for ; i < len(path)-1; i++ {
		if len(path) == 1 {
			break
		}
		if len(name) > 0 {
			name += "/"
		}
		name += path[i]
		if _, ok := f.items[name]; ok {
			break
		}
	}

	child := f.child(path[0], "", "")
	child.name = name

	// move ops with same path into new created folder
	ops := f.ops
	f.ops = nil
	pathName := child.path()
	for _, o := range ops {
		if o.Path.Path == pathName {
			child.ops = append(child.ops, o)
		} else {
			f.ops = append(f.ops, o)
		}
	}

	child.insertPath(path[i:], op)
}

func (f *brunoFolderBuilder) build(opt BrunoExportOptions) []any {
	var result []any
	seq := 1

	getHttpItemName := func(op *Operation, opt BrunoExportOptions) string {
		name := getBrunoItemName(op, opt)
		if opt.FolderArrangement == PathFolderArrangement {
			parent := f.path()
			name = strings.TrimPrefix(name, parent+"/")
			name = strings.Trim(name, "/")
		}
		return name
	}

	for _, name := range f.order {
		child := f.items[name]

		item := bruno.FolderItem{
			Info: &bruno.FolderInfo{
				Name:        child.name,
				Description: child.description,
				Type:        "folder",
				Sequence:    seq,
			},
			Items: child.build(opt),
		}
		if item.Info.Description == "" {
			item.Info.Description = child.summary
		}
		result = append(result, item)
		seq++
	}

	items := buildItems(f.ops, seq, opt, getHttpItemName)
	result = append(result, items...)

	return result
}

func (f *brunoFolderBuilder) path() string {
	if f.parent == nil {
		return "/"
	}
	parent := f.parent.path()
	return path.Join(parent, f.name)
}

func groupByTag(paths *PathItems, tagsByName map[string]*Tag, opt BrunoExportOptions) []any {
	root := &brunoFolderBuilder{items: map[string]*brunoFolderBuilder{}}
	var untagged []*Operation

	for it := paths.Iter(); it.Next(); {
		ref := it.Value()
		pathItem := ref.Value
		if pathItem == nil {
			continue
		}

		lookup := pathItem.Operations()
		for _, method := range pathItem.MethodOrder {
			op := lookup[method]
			if len(op.Tags) == 0 {
				untagged = append(untagged, op)
				continue
			}
			path := tagPath(tagsByName, op.Tags[0])
			if len(path) == 0 {
				// tag referenced but not declared in spec's top-level tags array
				untagged = append(untagged, op)
				continue
			}
			root.insertTag(path, op)
		}
	}

	items := root.build(opt)
	untaggedItems := buildItems(untagged, len(items)+1, opt, getBrunoItemName)
	items = append(items, untaggedItems...)

	return items
}

func groupByPath(paths *PathItems, opt BrunoExportOptions) []any {
	root := &brunoFolderBuilder{items: map[string]*brunoFolderBuilder{}}

	for it := paths.Iter(); it.Next(); {
		ref := it.Value()
		pathItem := ref.Value
		if pathItem == nil {
			continue
		}

		lookup := pathItem.Operations()
		for _, method := range pathItem.MethodOrder {
			op := lookup[method]

			path := strings.Split(op.Path.Path, "/")
			root.insertPath(path[1:], op)
		}
	}

	return root.build(opt)
}
