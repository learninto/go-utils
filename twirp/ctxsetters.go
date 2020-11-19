// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// ctxsetters is an implementation detail for twirp generated code, used
// by the generated servers to set values in contexts for later access with the
// twirp package's accessors.
//
// DO NOT USE CTXSETTERS OUTSIDE OF TWIRP'S GENERATED CODE.
package twirp

import (
	"context"
	"net/http"
	"strconv"

	"google.golang.org/protobuf/proto"
)

func WithMethodName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, MethodNameKey, name)
}

func WithServiceName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, ServiceNameKey, name)
}

func WithPackageName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, PackageNameKey, name)
}

func WithStatusCode(ctx context.Context, code int) context.Context {
	return context.WithValue(ctx, StatusCodeKey, strconv.Itoa(code))
}

func WithResponseWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, ResponseWriterKey, w)
}

func WithHttpRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, HttpRequestKey, r)
}

func WithRequest(ctx context.Context, r proto.Message) context.Context {
	return context.WithValue(ctx, RequestKey, r)
}

func WithResponse(ctx context.Context, resp proto.Message) context.Context {
	return context.WithValue(ctx, ResponseKey, resp)
}

func WithMethodOption(ctx context.Context, option string) context.Context {
	return context.WithValue(ctx, MethodOptionKey, option)
}
