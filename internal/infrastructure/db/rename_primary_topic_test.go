package db

import (
	"testing"

	atlas_schema "ariga.io/atlas/sql/schema"
	entschema "entgo.io/ent/dialect/sql/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubDiffer returns a fixed set of changes regardless of input, standing in
// for Ent's real differ so this test exercises only
// renamePostSentimentPrimaryTopicToCategory's own filtering logic.
type stubDiffer struct {
	changes []atlas_schema.Change
	err     error
}

func (s stubDiffer) Diff(_, _ *atlas_schema.Schema) ([]atlas_schema.Change, error) {
	return s.changes, s.err
}

// TestRenamePrimaryTopicToCategoryConvertsAddDropToRename verifies the hook
// detects the "add column category, drop column primary_topic" pair Ent's
// structural differ would normally produce for a renamed field, and replaces
// it with a single RenameColumn — so Atlas's executors (which already handle
// RenameColumn losslessly for both Postgres and SQLite) rename the column in
// place instead of Ent's default drop+recreate, which would lose the
// column's data.
func TestRenamePrimaryTopicToCategoryConvertsAddDropToRename(t *testing.T) {
	oldCol := &atlas_schema.Column{Name: "primary_topic", Type: &atlas_schema.ColumnType{Raw: "varchar(255)"}}
	newCol := &atlas_schema.Column{Name: "category", Type: &atlas_schema.ColumnType{Raw: "varchar(255)"}}
	table := &atlas_schema.Table{Name: "post_sentiments"}

	// An unrelated column change on the same table must survive untouched,
	// to confirm the hook only removes the specific pair it's targeting.
	unrelatedAdd := &atlas_schema.AddColumn{C: &atlas_schema.Column{Name: "feed_value_score"}}

	rawChanges := []atlas_schema.Change{
		&atlas_schema.ModifyTable{
			T: table,
			Changes: []atlas_schema.Change{
				&atlas_schema.AddColumn{C: newCol},
				unrelatedAdd,
				&atlas_schema.DropColumn{C: oldCol},
			},
		},
	}

	hook := renamePostSentimentPrimaryTopicToCategory()
	differ := hook(stubDiffer{changes: rawChanges})

	got, err := differ.Diff(nil, nil)
	require.NoError(t, err)
	require.Len(t, got, 1)

	mt, ok := got[0].(*atlas_schema.ModifyTable)
	require.True(t, ok, "expected a ModifyTable change")
	assert.Equal(t, "post_sentiments", mt.T.Name)

	require.Len(t, mt.Changes, 2, "add+drop pair must collapse into one rename, leaving the unrelated add untouched")

	var foundRename *atlas_schema.RenameColumn
	var foundUnrelated bool
	for _, c := range mt.Changes {
		switch v := c.(type) {
		case *atlas_schema.RenameColumn:
			foundRename = v
		case *atlas_schema.AddColumn:
			if v == unrelatedAdd {
				foundUnrelated = true
			}
		}
	}

	require.NotNil(t, foundRename, "expected a RenameColumn change in place of the add+drop pair")
	assert.Equal(t, oldCol, foundRename.From)
	assert.Equal(t, newCol, foundRename.To)
	assert.True(t, foundUnrelated, "unrelated AddColumn on the same table must be preserved")
}

// TestRenamePrimaryTopicToCategoryHandlesReverseRename verifies the hook is
// direction-agnostic: if the code's Ent schema is ever reverted back to
// `primary_topic`, the resulting "add column primary_topic, drop column
// category" pair is converted into a rename back to primary_topic too, not
// left to fall back on Ent's lossy default drop+recreate. This makes
// reverting the primary_topic->category rename itself safe, symmetric with
// the forward migration.
func TestRenamePrimaryTopicToCategoryHandlesReverseRename(t *testing.T) {
	oldCol := &atlas_schema.Column{Name: "category", Type: &atlas_schema.ColumnType{Raw: "varchar(255)"}}
	newCol := &atlas_schema.Column{Name: "primary_topic", Type: &atlas_schema.ColumnType{Raw: "varchar(255)"}}
	table := &atlas_schema.Table{Name: "post_sentiments"}

	rawChanges := []atlas_schema.Change{
		&atlas_schema.ModifyTable{
			T: table,
			Changes: []atlas_schema.Change{
				&atlas_schema.AddColumn{C: newCol},
				&atlas_schema.DropColumn{C: oldCol},
			},
		},
	}

	hook := renamePostSentimentPrimaryTopicToCategory()
	differ := hook(stubDiffer{changes: rawChanges})

	got, err := differ.Diff(nil, nil)
	require.NoError(t, err)
	require.Len(t, got, 1)

	mt, ok := got[0].(*atlas_schema.ModifyTable)
	require.True(t, ok, "expected a ModifyTable change")
	require.Len(t, mt.Changes, 1, "add+drop pair must collapse into one rename")

	rename, ok := mt.Changes[0].(*atlas_schema.RenameColumn)
	require.True(t, ok, "expected a RenameColumn change in place of the add+drop pair")
	assert.Equal(t, oldCol, rename.From)
	assert.Equal(t, newCol, rename.To)
}

// TestRenamePrimaryTopicToCategoryNoOpAfterRename verifies the hook does
// nothing (and errors on nothing) once the rename has already happened — the
// idempotency guarantee that lets it stay in the migration hook list
// permanently without a manual "run once" flag.
func TestRenamePrimaryTopicToCategoryNoOpAfterRename(t *testing.T) {
	// After the rename, a fresh diff against the live DB no longer has any
	// primary_topic/category add+drop pair to find — e.g. an unrelated
	// change on a different table, or no changes at all.
	rawChanges := []atlas_schema.Change{
		&atlas_schema.ModifyTable{
			T: &atlas_schema.Table{Name: "posts"},
			Changes: []atlas_schema.Change{
				&atlas_schema.AddColumn{C: &atlas_schema.Column{Name: "some_other_field"}},
			},
		},
	}

	hook := renamePostSentimentPrimaryTopicToCategory()
	differ := hook(stubDiffer{changes: rawChanges})

	got, err := differ.Diff(nil, nil)
	require.NoError(t, err)
	assert.Equal(t, rawChanges, got, "changes unrelated to the rename must pass through unmodified")
}

// TestRenamePrimaryTopicToCategoryPropagatesDifferError verifies the hook
// doesn't swallow an error from the wrapped differ.
func TestRenamePrimaryTopicToCategoryPropagatesDifferError(t *testing.T) {
	wantErr := assert.AnError

	hook := renamePostSentimentPrimaryTopicToCategory()
	differ := hook(stubDiffer{err: wantErr})

	_, err := differ.Diff(nil, nil)
	assert.ErrorIs(t, err, wantErr)
}

var _ entschema.Differ = stubDiffer{}
