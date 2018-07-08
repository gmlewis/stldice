#!/bin/bash -ex
go run gen-testdata.go g01234 # case g0 | g1 | g2 | g3 | g4:
go run gen-testdata.go g012346 # case g0 | g1 | g2 | g3 | g4 | g6:
go run gen-testdata.go g01235 # case g0 | g1 | g2 | g3 | g5:
go run gen-testdata.go g012357 # case g0 | g1 | g2 | g3 | g5 | g7:
go run gen-testdata.go g0124 # case g0 | g1 | g2 | g4:
go run gen-testdata.go g01245 # case g0 | g1 | g2 | g4 | g5:
go run gen-testdata.go g012457 # case g0 | g1 | g2 | g4 | g5 | g7:
go run gen-testdata.go g01246 # case g0 | g1 | g2 | g4 | g6:
go run gen-testdata.go g012467 # case g0 | g1 | g2 | g4 | g6 | g7:
go run gen-testdata.go g01247 # case g0 | g1 | g2 | g4 | g7:
go run gen-testdata.go g01256 # case g0 | g1 | g2 | g5 | g6:
go run gen-testdata.go g01257 # case g0 | g1 | g2 | g5 | g7:
go run gen-testdata.go g0126 # case g0 | g1 | g2 | g6:
go run gen-testdata.go g01267 # case g0 | g1 | g2 | g6 | g7:
go run gen-testdata.go g0127 # case g0 | g1 | g2 | g7:
go run gen-testdata.go g01345 # case g0 | g1 | g3 | g4 | g5:
go run gen-testdata.go g013456 # case g0 | g1 | g3 | g4 | g5 | g6:
go run gen-testdata.go g01346 # case g0 | g1 | g3 | g4 | g6:
go run gen-testdata.go g01347 # case g0 | g1 | g3 | g4 | g7:
go run gen-testdata.go g0135 # case g0 | g1 | g3 | g5:
go run gen-testdata.go g01356 # case g0 | g1 | g3 | g5 | g6:
go run gen-testdata.go g013567 # case g0 | g1 | g3 | g5 | g6 | g7:
go run gen-testdata.go g01357 # case g0 | g1 | g3 | g5 | g7:
go run gen-testdata.go g0136 # case g0 | g1 | g3 | g6:
go run gen-testdata.go g01367 # case g0 | g1 | g3 | g6 | g7:
go run gen-testdata.go g0137 # case g0 | g1 | g3 | g7:
go run gen-testdata.go g0146 # case g0 | g1 | g4 | g6:
go run gen-testdata.go g01467 # case g0 | g1 | g4 | g6 | g7:
go run gen-testdata.go g0147 # case g0 | g1 | g4 | g7:
go run gen-testdata.go g01567 # case g0 | g1 | g5 | g6 | g7:
go run gen-testdata.go g0157 # case g0 | g1 | g5 | g7:
go run gen-testdata.go g016 # case g0 | g1 | g6:
go run gen-testdata.go g017 # case g0 | g1 | g7:
go run gen-testdata.go g0234 # case g0 | g2 | g3 | g4:
go run gen-testdata.go g02345 # case g0 | g2 | g3 | g4 | g5:
go run gen-testdata.go g023456 # case g0 | g2 | g3 | g4 | g5 | g6:
go run gen-testdata.go g02346 # case g0 | g2 | g3 | g4 | g6:
go run gen-testdata.go g0235 # case g0 | g2 | g3 | g5:
go run gen-testdata.go g02356 # case g0 | g2 | g3 | g5 | g6:
go run gen-testdata.go g02357 # case g0 | g2 | g3 | g5 | g7:
go run gen-testdata.go g0236 # case g0 | g2 | g3 | g6:
go run gen-testdata.go g024 # case g0 | g2 | g4:
go run gen-testdata.go g0245 # case g0 | g2 | g4 | g5:
go run gen-testdata.go g02456 # case g0 | g2 | g4 | g5 | g6:
go run gen-testdata.go g0246 # case g0 | g2 | g4 | g6:
go run gen-testdata.go g02467 # case g0 | g2 | g4 | g6 | g7:
go run gen-testdata.go g0247 # case g0 | g2 | g4 | g7:
go run gen-testdata.go g025 # case g0 | g2 | g5:
go run gen-testdata.go g0256 # case g0 | g2 | g5 | g6:
go run gen-testdata.go g02567 # case g0 | g2 | g5 | g6 | g7:
go run gen-testdata.go g0257 # case g0 | g2 | g5 | g7:
go run gen-testdata.go g026 # case g0 | g2 | g6:
go run gen-testdata.go g0267 # case g0 | g2 | g6 | g7:
go run gen-testdata.go g027 # case g0 | g2 | g7:
go run gen-testdata.go g0345 # case g0 | g3 | g4 | g5:
go run gen-testdata.go g03456 # case g0 | g3 | g4 | g5 | g6:
go run gen-testdata.go g0346 # case g0 | g3 | g4 | g6:
go run gen-testdata.go g035 # case g0 | g3 | g5:
go run gen-testdata.go g0356 # case g0 | g3 | g5 | g6:
go run gen-testdata.go g0357 # case g0 | g3 | g5 | g7:
go run gen-testdata.go g036 # case g0 | g3 | g6:
go run gen-testdata.go g046 # case g0 | g4 | g6:
go run gen-testdata.go g0467 # case g0 | g4 | g6 | g7:
go run gen-testdata.go g056 # case g0 | g5 | g6:
go run gen-testdata.go g0567 # case g0 | g5 | g6 | g7:
go run gen-testdata.go g057 # case g0 | g5 | g7:
go run gen-testdata.go g067 # case g0 | g6 | g7:
go run gen-testdata.go g1234 # case g1 | g2 | g3 | g4:
go run gen-testdata.go g12345 # case g1 | g2 | g3 | g4 | g5:
go run gen-testdata.go g123457 # case g1 | g2 | g3 | g4 | g5 | g7:
go run gen-testdata.go g12346 # case g1 | g2 | g3 | g4 | g6:
go run gen-testdata.go g12347 # case g1 | g2 | g3 | g4 | g7:
go run gen-testdata.go g1235 # case g1 | g2 | g3 | g5:
go run gen-testdata.go g12356 # case g1 | g2 | g3 | g5 | g6:
go run gen-testdata.go g12357 # case g1 | g2 | g3 | g5 | g7:
go run gen-testdata.go g12367 # case g1 | g2 | g3 | g6 | g7:
go run gen-testdata.go g1237 # case g1 | g2 | g3 | g7:
go run gen-testdata.go g124 # case g1 | g2 | g4:
go run gen-testdata.go g1245 # case g1 | g2 | g4 | g5:
go run gen-testdata.go g12457 # case g1 | g2 | g4 | g5 | g7:
go run gen-testdata.go g1246 # case g1 | g2 | g4 | g6:
go run gen-testdata.go g1247 # case g1 | g2 | g4 | g7:
go run gen-testdata.go g1257 # case g1 | g2 | g5 | g7:
go run gen-testdata.go g1267 # case g1 | g2 | g6 | g7:
go run gen-testdata.go g127 # case g1 | g2 | g7:
go run gen-testdata.go g134 # case g1 | g3 | g4:
go run gen-testdata.go g1345 # case g1 | g3 | g4 | g5:
go run gen-testdata.go g13457 # case g1 | g3 | g4 | g5 | g7:
go run gen-testdata.go g1346 # case g1 | g3 | g4 | g6:
go run gen-testdata.go g13467 # case g1 | g3 | g4 | g6 | g7:
go run gen-testdata.go g1347 # case g1 | g3 | g4 | g7:
go run gen-testdata.go g135 # case g1 | g3 | g5:
go run gen-testdata.go g1356 # case g1 | g3 | g5 | g6:
go run gen-testdata.go g13567 # case g1 | g3 | g5 | g6 | g7:
go run gen-testdata.go g1357 # case g1 | g3 | g5 | g7:
go run gen-testdata.go g136 # case g1 | g3 | g6:
go run gen-testdata.go g1367 # case g1 | g3 | g6 | g7:
go run gen-testdata.go g137 # case g1 | g3 | g7:
go run gen-testdata.go g1457 # case g1 | g4 | g5 | g7:
go run gen-testdata.go g146 # case g1 | g4 | g6:
go run gen-testdata.go g1467 # case g1 | g4 | g6 | g7:
go run gen-testdata.go g147 # case g1 | g4 | g7:
go run gen-testdata.go g1567 # case g1 | g5 | g6 | g7:
go run gen-testdata.go g157 # case g1 | g5 | g7:
go run gen-testdata.go g167 # case g1 | g6 | g7:
go run gen-testdata.go g2345 # case g2 | g3 | g4 | g5:
go run gen-testdata.go g23456 # case g2 | g3 | g4 | g5 | g6:
go run gen-testdata.go g23457 # case g2 | g3 | g4 | g5 | g7:
go run gen-testdata.go g2346 # case g2 | g3 | g4 | g6:
go run gen-testdata.go g235 # case g2 | g3 | g5:
go run gen-testdata.go g2357 # case g2 | g3 | g5 | g7:
go run gen-testdata.go g245 # case g2 | g4 | g5:
go run gen-testdata.go g2456 # case g2 | g4 | g5 | g6:
go run gen-testdata.go g246 # case g2 | g4 | g6:
go run gen-testdata.go g2467 # case g2 | g4 | g6 | g7:
go run gen-testdata.go g247 # case g2 | g4 | g7:
go run gen-testdata.go g257 # case g2 | g5 | g7:
go run gen-testdata.go g345 # case g3 | g4 | g5:
go run gen-testdata.go g3456 # case g3 | g4 | g5 | g6:
go run gen-testdata.go g34567 # case g3 | g4 | g5 | g6 | g7:
go run gen-testdata.go g346 # case g3 | g4 | g6:
go run gen-testdata.go g356 # case g3 | g5 | g6:
go run gen-testdata.go g3567 # case g3 | g5 | g6 | g7:
go run gen-testdata.go g357 # case g3 | g5 | g7:
