package foundation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToUpperSnakeCase(t *testing.T) {

	t.Run("ReturnsLowercaseAsUppercase", func(t *testing.T) {

		// act
		snake := ToUpperSnakeCase("lowercase")

		assert.Equal(t, "LOWERCASE", snake)
	})

	t.Run("ReturnsPascalCaseAsUppercaseWithUnderscoreBetweenWords", func(t *testing.T) {

		// act
		snake := ToUpperSnakeCase("PascalCase")

		assert.Equal(t, "PASCAL_CASE", snake)
	})

	t.Run("ReturnsCamelCaseAsUppercaseWithUnderscoreBetweenWords", func(t *testing.T) {

		// act
		snake := ToUpperSnakeCase("camelCase")

		assert.Equal(t, "CAMEL_CASE", snake)
	})

	t.Run("ReturnsHyphenSeparatedCaseAsUppercaseWithUnderscoreBetweenWords", func(t *testing.T) {

		// act
		snake := ToUpperSnakeCase("kubernetes-engine")

		assert.Equal(t, "KUBERNETES_ENGINE", snake)
	})
}

func TestToLowerSnakeCase(t *testing.T) {

	t.Run("ReturnsLowercaseAsLowercase", func(t *testing.T) {

		// act
		snake := ToLowerSnakeCase("lowercase")

		assert.Equal(t, "lowercase", snake)
	})

	t.Run("ReturnsUppercaseAsLowercase", func(t *testing.T) {

		// act
		snake := ToLowerSnakeCase("LOWERCASE")

		assert.Equal(t, "lowercase", snake)
	})

	t.Run("ReturnsPascalCaseAsLowercaseWithUnderscoreBetweenWords", func(t *testing.T) {

		// act
		snake := ToLowerSnakeCase("PascalCase")

		assert.Equal(t, "pascal_case", snake)
	})

	t.Run("ReturnsCamelCaseAsLowercaseWithUnderscoreBetweenWords", func(t *testing.T) {

		// act
		snake := ToLowerSnakeCase("camelCase")

		assert.Equal(t, "camel_case", snake)
	})

	t.Run("ReturnsHyphenSeparatedCaseAsLowercaseWithUnderscoreBetweenWords", func(t *testing.T) {

		// act
		snake := ToLowerSnakeCase("kubernetes-engine")

		assert.Equal(t, "kubernetes_engine", snake)
	})
}

func TestFileExists(t *testing.T) {

	t.Run("ReturnsTrueIfFileExists", func(t *testing.T) {

		// act
		exists := FileExists("go.mod")

		assert.True(t, exists)
	})

	t.Run("ReturnsFalseIfFileDoesNotExist", func(t *testing.T) {

		// act
		exists := FileExists("go.pub")

		assert.False(t, exists)
	})

	t.Run("ReturnsFalseIfPathIsDirectory", func(t *testing.T) {

		// act
		exists := FileExists("..")

		assert.False(t, exists)
	})
}

func TestDirExists(t *testing.T) {

	t.Run("ReturnsTrueIfDirExists", func(t *testing.T) {

		// act
		exists := DirExists("..")

		assert.True(t, exists)
	})

	t.Run("ReturnsFalseIfDirDoesNotExist", func(t *testing.T) {

		// act
		exists := DirExists("vendor")

		assert.False(t, exists)
	})

	t.Run("ReturnsFalseIfPathIsFile", func(t *testing.T) {

		// act
		exists := DirExists("go.mod")

		assert.False(t, exists)
	})
}

func TestPathExists(t *testing.T) {

	t.Run("ReturnsTrueIfFileExists", func(t *testing.T) {

		// act
		exists := PathExists("go.mod")

		assert.True(t, exists)
	})

	t.Run("ReturnsTrueIfDirExists", func(t *testing.T) {

		// act
		exists := PathExists("..")

		assert.True(t, exists)
	})

	t.Run("ReturnsFalseIfDirDoesNotExist", func(t *testing.T) {

		// act
		exists := PathExists("vendor")

		assert.False(t, exists)
	})

	t.Run("ReturnsFalseIfFileDoesNotExist", func(t *testing.T) {

		// act
		exists := PathExists("go.pub")

		assert.False(t, exists)
	})
}
