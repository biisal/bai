package domain

import (
	"reflect"
	"testing"
)

func TestUnmarshalParts(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    []Part
		wantErr bool
	}{
		{
			name: "empty array",
			data: `[]`,
			want: []Part{},
		},
		{
			name: "text part",
			data: `[{"type":"text","data":{"Text":"hello"}}]`,
			want: []Part{
				{Type: PartTextType, Data: TextPartData{Text: "hello"}},
			},
		},
		{
			name: "reasoning part",
			data: `[{"type":"reasoning","data":{"Thinking":"think hard","Signature":"sig"}}]`,
			want: []Part{
				{Type: PartReasoningType, Data: ReasoningPartData{Thinking: "think hard", Signature: "sig"}},
			},
		},
		{
			name: "tool call part",
			data: `[{"type":"tool_call","data":{"ID":"call_1","Name":"read_file","Input":{"path":"/tmp/x"},"Finished":true}}]`,
			want: []Part{
				{
					Type: PartToolCallType,
					Data: ToolCallPartData{ID: "call_1", Name: "read_file", Input: []byte(`{"path":"/tmp/x"}`), Finished: true},
				},
			},
		},
		{
			name: "tool result part",
			data: `[{"type":"tool_result","data":{"ToolCallID":"call_1","Name":"read_file","Content":"file contents","IsError":false}}]`,
			want: []Part{
				{
					Type: PartToolResultType,
					Data: ToolResultPartData{ToolCallID: "call_1", Name: "read_file", Content: "file contents"},
				},
			},
		},
		{
			name: "finish part",
			data: `[{"type":"finish"}]`,
			want: []Part{
				{Type: PartFinishType},
			},
		},
		{
			name: "mixed parts",
			data: `[{"type":"text","data":{"Text":"a"}},{"type":"tool_call","data":{"ID":"c1"}}]`,
			want: []Part{
				{Type: PartTextType, Data: TextPartData{Text: "a"}},
				{Type: PartToolCallType, Data: ToolCallPartData{ID: "c1"}},
			},
		},
		{
			name:    "not json",
			data:    `hello world`,
			wantErr: true,
		},
		{
			name:    "unknown part type",
			data:    `[{"type":"mystery","data":{}}]`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalParts([]byte(tt.data))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("UnmarshalParts() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalParts() error = %v, want nil", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnmarshalParts() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
