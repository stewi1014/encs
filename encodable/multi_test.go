package encodable_test

import "github.com/stewi1014/encs/encodable"

var _ encodable.Encodable = &encodable.MultiAny{}
var _ encodable.Encodable = &encodable.MultiAll{}

// TODO: Write tests for multi
// Won't bother now as I'll have to re-write it later to use my integration testing framework that I haven't merged yet.
