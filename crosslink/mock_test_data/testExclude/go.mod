module go.opentelemetry.io/build-tools/crosslink/testroot

go 1.17

require go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0

// should not be replaced or overwritten
replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ../testA

// included in exclude slice and should be ignored
//replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB"

// should not be pruned
replace go.opentelemetry.io/build-tools/excludeme => ../excludeme
