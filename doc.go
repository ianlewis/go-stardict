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

// package stardict implements a library for reading stardict dictionaries in
// pure Go.
//
// Stardict dictionaries contain several files.
//
// * .ifo: The .ifo file contains metadata about the dictionary.
// * .idx: The .idx file is the dictionary index. It contains search entries
//   and associated offsets into the .dict file. The index file can be
//   compressed using gzip.
// * .dict: The .dict file is the dictionary's main article data. The dict file
//   can be compressed using the dictzip format.
// * .syn: The .syn file is optional and contains synonyms which link index
//   entries.
// * .tdx: The .tdx file is a list of word entries in a tree format. It can be
//   optionally present instead of the .idx file for tree dictionaries.
//
// More info on on the dictionary format can be found at this URL:
// https://github.com/huzheng001/stardict-3/blob/master/dict/doc/StarDictFileFormat
package stardict
