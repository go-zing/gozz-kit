package ztree

import (
	"bytes"
	"testing"
)

type (
	D struct {
		V *string
		s
	}

	s interface{}
)

var want = []byte(`digraph name {
node [style=filled shape=rect]
0 [fillcolor="#FBAED2" fontsize="16.0" height="0.25" label="[]interface{}" margin="0.05,0" tooltip="[]interface{}"];
0 -> 1 [arrowsize="0.7" fontsize="10" label="elements" labeltooltip="total: 2" weight="100"];
1 [fillcolor="#A6E7FF" fontsize="16.0" height="0.25" label="interface{}" margin="0.05,0" tooltip="interface{}"];
1 -> 2 [arrowhead="onormal" arrowsize="0.7" fontsize="10" label="implement" labeltooltip="any" weight="100"];
2 [fillcolor="#CCFF00" fontsize="16.0" height="0.25" label="ztree.D" margin="0.05,0" tooltip="github.com/go-zing/gozz-kit/ztree.D"];
subgraph cluster_2 {
style=dotted;
bgcolor="#f2fff2";
margin="0,0";
2 -> 3 [arrowhead="open" arrowsize="0.7" fontsize="10" label="V" labeltooltip="V" weight="100"];
2 -> 4 [arrowhead="open" arrowsize="0.7" fontsize="10" label="s" labeltooltip="s" weight="100"];
}
3 [fillcolor="#FBE7B2" fontsize="16.0" height="0.25" label="string" margin="0.05,0" tooltip="string"];
4 [fillcolor="#A6E7FF" fontsize="8.0" height="0.25" label="ztree.s" margin="0.05,0" tooltip="github.com/go-zing/gozz-kit/ztree.s"];
4 -> 3 [arrowhead="onormal" arrowsize="0.7" fontsize="10" label="implement" labeltooltip="any" style="dashed" weight="100"];
1 -> 6 [arrowhead="onormal" arrowsize="0.7" fontsize="10" label="implement" labeltooltip="any" weight="100"];
6 [fillcolor="#CCFF00" fontsize="16.0" height="0.25" label="ztree.D" margin="0.05,0" tooltip="github.com/go-zing/gozz-kit/ztree.D"];
subgraph cluster_6 {
style=dotted;
bgcolor="#f2fff2";
margin="0,0";
6 -> 3 [arrowhead="open" arrowsize="0.7" fontsize="10" label="V" labeltooltip="V" weight="100"];
6 -> 7 [arrowhead="open" arrowsize="0.7" fontsize="10" label="s" labeltooltip="s" weight="100"];
}
7 [fillcolor="#A6E7FF" fontsize="8.0" height="0.25" label="ztree.s" margin="0.05,0" tooltip="github.com/go-zing/gozz-kit/ztree.s"];
7 -> 8 [arrowhead="onormal" arrowsize="0.7" fontsize="10" label="implement" labeltooltip="any" style="dashed" weight="100"];
8 [fillcolor="#FBE7B2" fontsize="8.0" height="0.25" label="string" margin="0.05,0" tooltip="string"];
}`)

func TestParseAndDraw(t *testing.T) {
	ptr := new(string)
	d1 := &D{V: ptr, s: ptr}
	if !bytes.Equal(want, (&Drawer{}).Draw(Parse([]interface{}{d1, D{
		V: ptr,
		s: ptr,
	}}, WithUnexported(true)))) {
		t.Fatal()
	}
}
