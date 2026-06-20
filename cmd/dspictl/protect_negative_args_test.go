package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("protectNegativeArgs", func() {
	var (
		input  []string
		result []string
	)

	JustBeforeEach(func() {
		result = protectNegativeArgs(input)
	})

	Context("with a negative number after a known value-taking flag", func() {
		Context("with --gain", func() {
			BeforeEach(func() {
				input = []string{"eq", "master", "set", "0", "0", "--freq", "64.8", "--gain", "-8.9", "--q", "6.039", "--type", "peak"}
			})

			It("leaves the argument list unchanged", func() {
				Expect(result).To(Equal([]string{"eq", "master", "set", "0", "0", "--freq", "64.8", "--gain", "-8.9", "--q", "6.039", "--type", "peak"}))
			})
		})

		Context("with --freq", func() {
			BeforeEach(func() {
				input = []string{"eq", "master", "set", "0", "0", "--freq", "-64.8", "--gain", "8.9", "--q", "6.039", "--type", "peak"}
			})

			It("leaves the argument list unchanged", func() {
				Expect(result).To(Equal([]string{"eq", "master", "set", "0", "0", "--freq", "-64.8", "--gain", "8.9", "--q", "6.039", "--type", "peak"}))
			})
		})

		Context("with --q", func() {
			BeforeEach(func() {
				input = []string{"eq", "master", "set", "0", "0", "--freq", "64.8", "--gain", "8.9", "--q", "-6.039", "--type", "peak"}
			})

			It("leaves the argument list unchanged", func() {
				Expect(result).To(Equal([]string{"eq", "master", "set", "0", "0", "--freq", "64.8", "--gain", "8.9", "--q", "-6.039", "--type", "peak"}))
			})
		})
	})

	Context("with a positive number after a known flag", func() {
		BeforeEach(func() {
			input = []string{"eq", "master", "set", "0", "0", "--freq", "64.8", "--gain", "8.9", "--q", "6.039", "--type", "peak"}
		})

		It("leaves the argument list unchanged", func() {
			Expect(result).To(Equal([]string{"eq", "master", "set", "0", "0", "--freq", "64.8", "--gain", "8.9", "--q", "6.039", "--type", "peak"}))
		})
	})

	Context("with a negative number not preceded by a known flag", func() {
		BeforeEach(func() {
			input = []string{"volume", "set", "-20"}
		})

		It("inserts -- before the negative number", func() {
			Expect(result).To(Equal([]string{"volume", "set", "--", "-20"}))
		})
	})

	Context("with a negative number after an explicit --", func() {
		BeforeEach(func() {
			input = []string{"cmd", "--", "-20"}
		})

		It("leaves the argument list unchanged", func() {
			Expect(result).To(Equal([]string{"cmd", "--", "-20"}))
		})
	})

	Context("with multiple negative numbers without preceding flags", func() {
		BeforeEach(func() {
			input = []string{"a", "-1", "b", "-2.5"}
		})

		It("inserts -- before each negative number", func() {
			Expect(result).To(Equal([]string{"a", "--", "-1", "b", "--", "-2.5"}))
		})
	})
})
