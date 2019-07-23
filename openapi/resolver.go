package openapi

import "fmt"

type Resolver struct {
	components *Components
	antiCircle map[Reference]interface{}
}

type refVisitor func(r *Resolver) error

func (r *Resolver) visit(ref Reference, f refVisitor) error {
	if ref.Ref == "" {
		return nil
	}

	_, visited := r.antiCircle[ref]
	if visited {
		return nil
	}

	r.antiCircle[ref] = nil
	defer delete(r.antiCircle, ref)
	return f(r)
}

func (api *OpenAPI) resolveReferences() error {
	r := &Resolver{
		components: &api.Components,
		antiCircle: make(map[Reference]interface{}),
	}

	err := api.Paths.visit(r)
	if err != nil {
		return err
	}

	return nil
}

func (paths *Paths) visit(r *Resolver) error {
	for _, pathEntry := range paths.paths {
		if pathEntry.path != nil {
			err := pathEntry.path.visit(r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (path *Path) visit(r *Resolver) error {
	for _, method := range []*Method{
		path.Get,
		path.Post,
		path.Put,
		path.Patch,
		path.Delete,
		path.Head,
		path.Options,
		path.Trace,
	} {
		if method != nil {
			err := method.visit(r)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Method) visit(r *Resolver) error {
	for _, p := range m.Parameters {
		err := r.visit(p.Reference, p.visit)
		if err != nil {
			return err
		}

		err = r.visit(p.Schema.Reference, p.Schema.visit)
		if err != nil {
			return err
		}
	}

	for _, resp := range m.Responses {
		err := resp.visit(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Parameter) visit(r *Resolver) error {
	v, ok := r.components.Parameters[p.Reference.RefName()]
	if !ok {
		return fmt.Errorf("failed to resolve %s", p.Ref)
	}
	p.ParameterImpl = v.ParameterImpl
	return nil
}

func (s *Schema) visit(r *Resolver) error {
	v, ok := r.components.Schemas[s.Reference.RefName()]
	if !ok {
		return fmt.Errorf("failed to resolve %s", s.Ref)
	}
	s.SchemaImpl = v.SchemaImpl
	return nil
}

func (s *Response) visit(r *Resolver) error {
	if s.Ref != "" {
		v, ok := r.components.Responses[s.Reference.RefName()]
		if !ok {
			return fmt.Errorf("failed to resolve %s", s.Ref)
		}
		s.ResponseImpl = v.ResponseImpl
	}

	for _, cont := range s.Content {
		if cont.Schema != nil {
			err := r.visit(cont.Schema.Reference, cont.Schema.visit)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
