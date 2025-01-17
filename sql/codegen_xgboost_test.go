// Copyright 2019 The SQLFlow Authors. All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sql

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testXGTrainSelectIris = `
SELECT *
FROM iris.train
TRAIN xgboost.Estimator
WITH
	train.objective = "multi:softmax",
	train.num_class = 3,
	train.max_depth = 5,
	train.eta = 0.3,
	train.tree_method = "approx",
	train.num_round = 30
COLUMN sepal_length, sepal_width, petal_length, petal_width
LABEL class INTO sqlflow_models.my_xgboost_model;
`

	testXGPredSelectIris = `
SELECT *
FROM iris.test
PREDICT iris.predict
WITH
	pred.append_columns = [sepal_length, sepal_width, petal_length, petal_width],
	pred.prob_column = prob,
	pred.detail_column = detail,
	pred.encoding_column = encoding
USING sqlflow_models.my_xgboost_model;
`
)

func TestPartials(t *testing.T) {
	a := assert.New(t)
	tmpMap := make(map[string][]string)
	filler := &xgboostFiller{}

	// test strPartial
	part := strPartial("obj", func(r *xgboostFiller) *string { return &(r.Objective) })
	tmpMap["obj"] = []string{"binary:logistic"}
	e := part(&tmpMap, filler)
	a.NoError(e)
	a.Equal(filler.Objective, "binary:logistic")
	_, ok := tmpMap["obj"]
	a.Equal(ok, false)
	// Error: duplicate attr setting
	tmpMap["obj"] = []string{"binary:logistic"}
	e = part(&tmpMap, filler)
	a.Error(e)
	// Error: len(val) > 1
	tmpMap["obj"] = []string{"binary:logistic", "reg:linear"}
	e = part(&tmpMap, filler)
	a.Error(e)
	//  change objective to "reg:linear"
	tmpMap["obj"] = []string{"reg:linear"}
	filler.Objective = ""
	e = part(&tmpMap, filler)
	a.NoError(e)
	a.Equal(filler.Objective, "reg:linear")

	// test uIntPartial
	part = uIntPartial("num_class", func(r *xgboostFiller) *uint { return &(r.NumClass) })
	tmpMap["num_class"] = []string{"3"}
	e = part(&tmpMap, filler)
	a.NoError(e)
	a.EqualValues(filler.NumClass, 3)
	_, ok = tmpMap["num_class"]
	a.Equal(ok, false)

	// test fp32Partial
	part = fp32Partial("eta", func(r *xgboostFiller) *float32 { return &(r.Eta) })
	tmpMap["eta"] = []string{"-0.33"}
	e = part(&tmpMap, filler)
	a.NoError(e)
	a.EqualValues(filler.Eta, float32(-0.33))
	_, ok = tmpMap["eta"]
	a.Equal(ok, false)

	// test boolPartial
	part = boolPartial("auto_train", func(r *xgboostFiller) *bool { return &(r.AutoTrain) })
	tmpMap["auto_train"] = []string{"false"}
	e = part(&tmpMap, filler)
	a.NoError(e)
	a.Equal(filler.AutoTrain, false)
	_, ok = tmpMap["auto_train"]
	a.Equal(ok, false)
	tmpMap["auto_train"] = []string{"true"}
	e = part(&tmpMap, filler)
	a.NoError(e)
	a.Equal(filler.AutoTrain, true)

	// test sListPartial
	part = sListPartial("append_columns", func(r *xgboostFiller) *[]string { return &(r.AppendColumns) })
	tmpMap["append_columns"] = []string{"AA", "BB", "CC"}
	e = part(&tmpMap, filler)
	a.NoError(e)
	a.EqualValues(filler.AppendColumns, []string{"AA", "BB", "CC"})
	_, ok = tmpMap["append_columns"]
	a.Equal(ok, false)
}

func TestXGBoostAttr(t *testing.T) {
	a := assert.New(t)
	assertEq := func(m map[string]interface{}, key string, refVal interface{}) {
		val, _ := m[key]
		a.EqualValues(refVal, val)
	}
	parser := newParser()

	parseAndFill := func(clause string) *xgboostFiller {
		filler := &xgboostFiller{}
		r, e := parser.Parse(clause)
		a.NoError(e)
		e = xgParseAttr(r, filler)
		a.NoError(e)
		return filler
	}

	trainClause := `
SELECT a, b, c, d, e FROM table_xx
TRAIN xgboost.Estimator
WITH
	train.objective = "binary:logistic",
	train.booster = gblinear,
	train.num_class = 2,
	train.max_depth = 5,
	train.eta = 0.03,
	train.tree_method = hist,
	train.subsample = 0.8,
	train.colsample_bytree = 0.5,
	train.colsample_bylevel = 0.6,
	train.max_bin = 128,
	train.verbosity = 3,
	train.num_round = 300,
	train.auto_train = true
COLUMN a, b, c, d
LABEL e INTO table_123;
`
	filler := parseAndFill(trainClause)
	data, e := json.Marshal(filler.xgLearningFields)
	a.NoError(e)
	mapData := make(map[string]interface{})
	e = json.Unmarshal(data, &mapData)
	a.NoError(e)
	params, _ := mapData["params"]
	paramMap, _ := params.(map[string]interface{})
	assertEq(paramMap, "objective", "binary:logistic")
	assertEq(paramMap, "booster", "gblinear")
	assertEq(paramMap, "num_class", 2)
	assertEq(paramMap, "max_depth", 5)
	assertEq(paramMap, "eta", 0.03)
	assertEq(paramMap, "tree_method", "hist")
	assertEq(paramMap, "subsample", 0.8)
	assertEq(paramMap, "colsample_bytree", 0.5)
	assertEq(paramMap, "colsample_bylevel", 0.6)
	assertEq(paramMap, "max_bin", 128)
	assertEq(paramMap, "verbosity", 3)
	assertEq(mapData, "num_boost_round", 300)
	assertEq(mapData, "auto_train", true)

	predClause := `
SELECT a, b, c, d, e FROM table_xx
PREDICT table_yy
WITH
	pred.detail_column = "prediction_detail",
	pred.prob_column = "prediction_probability",
	pred.encoding_column = "prediction_leafs",
	pred.result_column = "prediction_results",
	pred.append_columns = ["AA", "BB", "CC"]
USING sqlflow_models.my_xgboost_model;
`
	filler = parseAndFill(predClause)
	a.EqualValues([]string{"AA", "BB", "CC"}, filler.AppendColumns)
	a.EqualValues("prediction_probability", filler.ProbColumn)
	a.EqualValues("prediction_results", filler.ResultColumn)
	a.EqualValues("prediction_detail", filler.DetailColumn)
	a.EqualValues("prediction_leafs", filler.EncodingColumn)
}

func TestColumnClause(t *testing.T) {
	a := assert.New(t)
	parser := newParser()
	sqlHead := `
SELECT a, b, c, d, e FROM table_xx
TRAIN xgboost.Estimator
WITH attr_x = XXX
`
	sqlTail := `
LABEL e INTO model_table;
`
	// test sparseKV schema
	filler := &xgboostFiller{}
	sparseKVSpec := ` COLUMN SPARSE(a, 100, comma) `
	r, e := parser.Parse(sqlHead + sparseKVSpec + sqlTail)
	a.NoError(e)
	e = xgParseColumns(r, filler)
	a.NoError(e)
	a.EqualValues(100, filler.FeatureSize)
	a.EqualValues(",", filler.Delimiter)
	a.EqualValues(true, filler.IsSparse)
	a.EqualValues([]string{"a"}, filler.FeatureColumns)
	a.EqualValues("a", filler.X[0].FeatureName)
	a.EqualValues("string", filler.X[0].Dtype)
	a.EqualValues("", filler.X[0].Delimiter)
	a.EqualValues("", filler.X[0].InputShape)
	a.EqualValues(false, filler.X[0].IsSparse)
	a.EqualValues("", filler.X[0].FeatureColumnCode)
	a.EqualValues(false, filler.IsTensorFlowIntegrated)
	a.EqualValues(&xgFeatureMeta{FeatureName: "e"}, filler.LabelField)
	a.EqualValues("e", filler.Label)

	// test raw columns
	filler = &xgboostFiller{}
	rawColumnsSpec := " COLUMN a, b, b, c, d, c "
	r, _ = parser.Parse(sqlHead + rawColumnsSpec + sqlTail)
	e = xgParseColumns(r, filler)
	a.NoError(e)
	a.EqualValues(6, int(filler.FeatureSize))
	a.EqualValues("", filler.Delimiter)
	a.False(filler.IsSparse)
	a.False(filler.IsTensorFlowIntegrated)
	feaKeys := []string{"a", "b", "b", "c", "d", "c"}
	a.EqualValues(feaKeys, filler.FeatureColumns)
	for i, key := range feaKeys {
		a.EqualValues(key, filler.X[i].FeatureName)
		a.EqualValues("float32", filler.X[i].Dtype)
		a.EqualValues("", filler.X[i].Delimiter)
		a.EqualValues("[1]", filler.X[i].InputShape)
		a.EqualValues(false, filler.X[i].IsSparse)
		a.EqualValues("", filler.X[i].FeatureColumnCode)
	}

	// test tf.feature_columns
	filler = &xgboostFiller{}
	fcSpec := " COLUMN a, b, c, EMBEDDING(CATEGORY_ID(d, 2000), 8, mean) FOR feature_columns "
	r, _ = parser.Parse(sqlHead + fcSpec + sqlTail)
	e = xgParseColumns(r, filler)
	a.NoError(e)
	a.EqualValues(0, int(filler.FeatureSize))
	a.EqualValues("", filler.Delimiter)
	a.False(filler.IsSparse)
	a.True(filler.IsTensorFlowIntegrated)

	// test group & weight
	filler = &xgboostFiller{}
	groupWeightSpec := " COLUMN gg FOR group COLUMN ww FOR weight "
	r, _ = parser.Parse(sqlHead + fcSpec + groupWeightSpec + sqlTail)
	e = xgParseColumns(r, filler)
	a.NoError(e)
	a.EqualValues(&xgFeatureMeta{FeatureName: "gg"}, filler.GroupField)
	a.EqualValues("gg", filler.Group)
	a.EqualValues(&xgFeatureMeta{FeatureName: "ww"}, filler.WeightField)
	a.EqualValues("ww", filler.Weight)

	// test xgMixSchemaError
	filler = &xgboostFiller{}
	wrongColSpec := " COLUMN SPARSE(a, 2000, comma), b, c, d "
	r, _ = parser.Parse(sqlHead + wrongColSpec + sqlTail)
	e = xgParseColumns(r, filler)
	a.Error(e)
	a.EqualValues(e, xgParseColumnError("feature_columns", xgMixSchemaError()))

	// test `DENSE` keyword
	filler = &xgboostFiller{}
	wrongColSpec = " COLUMN DENSE(b, 5, comma) "
	r, _ = parser.Parse(sqlHead + wrongColSpec + sqlTail)
	e = xgParseColumns(r, filler)
	a.Error(e)
	a.EqualValues(e, xgParseColumnError("feature_columns", xgUnknownFCError("DENSE")))

	// test xgMultiSparseError
	filler = &xgboostFiller{}
	wrongColSpec = " COLUMN SPARSE(a, 2000, comma), SPARSE(b, 100, comma) "
	r, _ = parser.Parse(sqlHead + wrongColSpec + sqlTail)
	e = xgParseColumns(r, filler)
	a.Error(e)
	a.EqualValues(e, xgParseColumnError("feature_columns", xgMultiSparseError([]string{"a", "b"})))

	// test xgUnsupportedColTagError
	filler = &xgboostFiller{}
	unsupportedSpec := " COLUMN gg FOR group COLUMN ww FOR xxxxx "
	r, _ = parser.Parse(sqlHead + fcSpec + unsupportedSpec + sqlTail)
	e = xgParseColumns(r, filler)
	a.Error(e)
	a.EqualValues(e, xgParseColumnError("xxxxx", xgUnsupportedColTagError()))
}

func TestXGBoostFiller(t *testing.T) {
	a := assert.New(t)

	parser := newParser()
	trainClause := `
SELECT * FROM iris.train
TRAIN xgboost.Regressor
WITH
	train.max_depth = 5,
	train.eta = 0.03,
	train.tree_method = "hist",
	train.num_round = 300
COLUMN sepal_length, sepal_width, petal_length, petal_width
COLUMN gg FOR group 
COLUMN ww FOR weight
LABEL e INTO model_table;
`
	pr, e := parser.Parse(trainClause)
	a.NoError(e)
	fts, e := verify(pr, testDB)
	a.NoError(e)
	filler, e := newXGBoostFiller(pr, fts, testDB)
	a.NoError(e)

	a.True(filler.IsTrain)
	stdSlct := removeLastSemicolon(strings.Replace(filler.StandardSelect, "\n", " ", -1))
	a.EqualValues("SELECT * FROM iris.train", stdSlct)
	a.EqualValues("model_table", filler.ModelPath)

	a.EqualValues("reg:squarederror", filler.Objective)
	a.EqualValues(0.03, filler.Eta)
	a.EqualValues(5, filler.MaxDepth)
	a.EqualValues("hist", filler.TreeMethod)
	a.EqualValues(300, filler.NumRound)

	a.EqualValues("e", filler.Label)
	a.EqualValues("e", filler.LabelField.FeatureName)
	a.EqualValues("gg", filler.Group)
	a.EqualValues("gg", filler.GroupField.FeatureName)
	a.EqualValues("ww", filler.Weight)
	a.EqualValues("ww", filler.WeightField.FeatureName)

	a.False(filler.IsTensorFlowIntegrated)
	a.False(filler.IsSparse)
	a.EqualValues("", filler.Delimiter)
	a.EqualValues(4, filler.FeatureSize)
	a.EqualValues([]string{"sepal_length", "sepal_width", "petal_length", "petal_width"}, filler.FeatureColumns)
	a.EqualValues(&xgFeatureMeta{FeatureName: "sepal_length", Dtype: "float32", InputShape: "[1]"}, filler.X[0])
	a.EqualValues(&xgFeatureMeta{FeatureName: "sepal_width", Dtype: "float32", InputShape: "[1]"}, filler.X[1])
	a.EqualValues(&xgFeatureMeta{FeatureName: "petal_length", Dtype: "float32", InputShape: "[1]"}, filler.X[2])
	a.EqualValues(&xgFeatureMeta{FeatureName: "petal_width", Dtype: "float32", InputShape: "[1]"}, filler.X[3])

	colFields := &xgColumnFields{}
	e = json.Unmarshal([]byte(filler.ColumnJSON), colFields)
	a.NoError(e)
	a.EqualValues(filler.xgColumnFields, *colFields)
	dsFields := &xgDataSourceFields{}
	e = json.Unmarshal([]byte(filler.DataSourceJSON), dsFields)
	a.NoError(e)
	a.EqualValues(filler.xgDataSourceFields, *dsFields)
	xgbFields := &xgLearningFields{}
	e = json.Unmarshal([]byte(filler.LearningJSON), xgbFields)
	a.NoError(e)
	a.EqualValues(filler.xgLearningFields, *xgbFields)
}
