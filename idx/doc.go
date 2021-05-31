// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package idx implements reading .idx files.
//
// The .idx file contains a list of dictionary entry titles and the associated
// offset and size of the main article content in the .dict file.
//
// Each .idx file entry (word) comes in three parts:
//   1. The title: a utf-8 string terminated by a null terminator ('\0').
//   2. The offset: a 32 or 64 bit integer offset of the word in the .dict
//      file in network byte order.
//   3. The size: a 32 bit integer size of the word in the .dict file in
//      network byte order.
package idx
