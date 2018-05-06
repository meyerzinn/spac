//go:generate flatc --gen-all --go-namespace upstream --go protocol/upstream/upstream.fbs
//go:generate flatc --gen-all --go-namespace downstream --go protocol/downstream/downstream.fbs
package net
