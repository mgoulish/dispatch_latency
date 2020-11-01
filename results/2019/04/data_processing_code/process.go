package main

import ( "fmt"  
         "os"
         "os/exec"
         "io/ioutil"
         "math"
         "path/filepath"
         "sort"
         "strings"
         "strconv"
       )



var fp = fmt.Fprintf 





type latency_1_result struct {
  name               string
  file_names      [] string
  flight_times    [] float64
  max_flight_time    float64
  bin_size           float64
  bins            [] int
  n_nines            int
  nines_positions [] int
  nines_values    [] float64
  y_max_1            int
  y_max_2            int
}





func new_latency_1_result ( name string ) ( * latency_1_result ) {
  l1r                 := & latency_1_result { name : name }
  l1r.n_nines         = 6
  l1r.nines_positions = make ( [ ] int,     l1r.n_nines )
  l1r.nines_values    = make ( [ ] float64, l1r.n_nines )

  return l1r
}





func ( l1r * latency_1_result ) get_files ( dir string, substr string ) {

  _ = filepath.Walk ( dir, 
                      func ( path string, info os.FileInfo, err error) error {
                        if ! info.IsDir ( ) {
                          if strings.Contains ( path, substr ) {
                            l1r.file_names = append ( l1r.file_names, path )
                          }
                        }
                        return nil
                      } )
}





func ( l1r * latency_1_result ) get_floats_from_file ( path string ) {

  content, err := ioutil.ReadFile ( path )
  if err != nil {
    fmt.Println ( "dang." );
    os.Exit ( 1 );
  }
  lines := strings.Split ( string(content), "\n" )

  for _, line := range lines {
    num, err := strconv.ParseFloat ( line, 64 )
    if err == nil {
      l1r.flight_times = append ( l1r.flight_times, num )
    }
  }
}





func ( l1r * latency_1_result ) get_flight_times ( dir string ) {

  l1r.get_files ( dir, "flight_times" )

  for _, file := range l1r.file_names {
    l1r.get_floats_from_file ( file )
  }
}





func ( l1r * latency_1_result ) convert_to_msec ( ) {
  for i := 0; i < len ( l1r.flight_times ); i ++ {
    l1r.flight_times [ i ] = l1r.flight_times [ i ] * 1000
  }
}





func ( l1r * latency_1_result ) find_max ( ) ( ) {
  l1r.max_flight_time = - math.MaxFloat64
  for _, n := range l1r.flight_times {
    if n > l1r.max_flight_time {
      l1r.max_flight_time = n
    }
  }
}





func ( l1r * latency_1_result ) bin ( ) {

  // This bin size will be applied to the flight times after 
  // they have been converted to msec, so it is 0.01 msec == 10 usec.
  l1r.bin_size = 0.01
  n_bins := 1 + int ( l1r.max_flight_time / l1r.bin_size )

  l1r.bins = make ( [] int, n_bins )

  for _, n := range l1r.flight_times {
    bin := int ( n / l1r.bin_size )
    l1r.bins [ bin ] ++
  }
}





func ( l1r * latency_1_result ) nines ( ) {
  sort.Float64s ( l1r.flight_times )
  n_numbers := len ( l1r.flight_times )

  the_next_nine  := 0.9
  nines_fraction := 0.0

  for nine := 0; nine < l1r.n_nines; nine ++ {
    nines_fraction += the_next_nine
    l1r.nines_positions [ nine ] = int(nines_fraction * float64(n_numbers))
    the_next_nine /= 10
  }

  for nine := 0; nine < l1r.n_nines; nine ++ {
    position := l1r.nines_positions [ nine ]
    nine_value := l1r.flight_times [ position ]
    fp ( os.Stdout, "%d nines : %.7f\n", nine + 1, nine_value )
    l1r.nines_values [ nine ] = nine_value
  }
}





func ( l1r * latency_1_result ) find_max_to_2_nines ( ) {
  two_nines_value := l1r.nines_values [ 1 ]
  two_nines_bin := int ( two_nines_value / l1r.bin_size )

  // Look at bin values (which will be Y-value on the graph)
  // and find the highest one in the first part of the bins,
  // from 0 to 2-nines.
  l1r.y_max_1 = 0
  for i := 0; i < two_nines_bin; i ++ {
    if l1r.bins[i] > l1r.y_max_1 {
      l1r.y_max_1 = l1r.bins[i]
    }
  }

  // Now find the highest bin value from 2-nines to the end.
  l1r.y_max_2 = 0
  for i := two_nines_bin; i < len ( l1r.bins ); i ++ {
    if l1r.bins[i] > l1r.y_max_2 {
      l1r.y_max_2 = l1r.bins[i]
    }
  }
}





func ( l1r * latency_1_result ) process_latency_1_data ( dir string ) {
  l1r.get_flight_times ( dir + "/" + l1r.name )
  l1r.convert_to_msec ( )
  l1r.find_max ( )

  l1r.bin ( )
  l1r.nines ( )
  l1r.find_max_to_2_nines ( )
}





func ( l1r * latency_1_result ) run_gnuplot ( ) {

  number_words :=  [ 6 ] string { "ONE", "TWO", "THREE", "FOUR", "FIVE", "SIX" }

  if l1r.n_nines > 6 {
    fp ( os.Stdout, "Too Many Nines!\n" )
    os.Exit ( 1 )
  }

  // gps means "gnuplot script"

  // Some boilerplate instructions to gnuplot, that 
  // are always the same.
  gps := "unset key\n"
  gps += "set xlabel \"latency (msec)\"\n"
  gps += "set ylabel \"number of messages\"\n"
  gps += "set terminal jpeg size 1200,500\n"
  gps += "set style textbox opaque noborder\n"

  // These are variables that indicate the maximum Y values for the first and second
  // graph.  We need these so we can place the nines-labels properly. 
  gps += "\n"
  gps += fmt.Sprintf ( "Y_MAX_1 = %d\n", l1r.y_max_1 )
  gps += fmt.Sprintf ( "Y_MAX_2 = %d\n", l1r.y_max_2 )
  gps += "\n"

  for i := 0; i < l1r.n_nines; i ++ {
    gps += fmt.Sprintf ( "%s_NINE    = %.7f\n", number_words[i], l1r.nines_values [ i ] )
  }

  gps += "\n"

  // Create the first two arrows, which will appear on the first graph.
  // These are the red vertical lines that line up beneath the nines labels.
  for i := 0; i < 2; i ++ {
    gps += fmt.Sprintf ( "set arrow %d from %s_NINE,    0 to %s_NINE,    Y_MAX_1 nohead back nofilled lc rgb \"red\" lw 3.000 dashtype solid\n", i + 1, number_words[i], number_words[i] )
  }

  // Create the last four arrows, which will appear on the second graph.
  // These are the red vertical lines that line up beneath the nines labels.
  for i := 2; i < l1r.n_nines; i ++ {
    gps += fmt.Sprintf ( "set arrow %d from %s_NINE,    0 to %s_NINE,    Y_MAX_2 nohead back nofilled lc rgb \"red\" lw 3.000 dashtype solid\n", i + 1, number_words[i], number_words[i] )
  }

  // Create the labels for the first two nines, which will appear in the first graph.
  gps += "\n"
  gps += "LABEL_Y=Y_MAX_1\n" // The Y value of the nines labels will be near the top of this graph.
  gps += "\n"

  for i := 0; i < 2; i ++ {
    gps += fmt.Sprintf ( "LABEL = \"%d Nine\"\n", i + 1 )
    gps += fmt.Sprintf ( "set obj %d rect at %s_NINE , LABEL_Y size char strlen(LABEL), char 1 fc rgb \"white\" front\n", i + 1, number_words[i] )
    gps += fmt.Sprintf ( "set label %d LABEL at %s_NINE , LABEL_Y front center\n", i + 1, number_words[i] )
    gps += "\n"
  }

  gps += "\n"
  gps += "LABEL_Y=Y_MAX_2\n" // The Y value of the nines labels will be near the top of *this* graph.
  gps += "\n"

  // Create the labels for the last four nines, which will appear in the second graph.
  for i := 2; i < l1r.n_nines; i ++ {
    gps += fmt.Sprintf ( "LABEL = \"%d Nine\"\n", i + 1 )
    gps += fmt.Sprintf ( "set obj %d rect at %s_NINE , LABEL_Y size char strlen(LABEL), char 1 fc rgb \"white\" front\n", i + 1, number_words[i] )
    gps += fmt.Sprintf ( "set label %d LABEL at %s_NINE , LABEL_Y front center\n", i + 1, number_words[i] )
    gps += "\n"
  }


  // BUGALERT -- this is embedded knowledge of how the test works.
  // May not always be true.
  // And of how the test name is assigned.
  n_receivers, _ := strconv.Atoi ( l1r.name )
  n_senders      := n_receivers * 40

  millions_of_messages := float64(len(l1r.flight_times))/1000000.0

  // Do the plotting clause for the first graph.
  gps += fmt.Sprintf ( "set output \"%s_receivers.jpg\"\n", l1r.name )
  gps += "set xrange [ 0 : TWO_NINE ]\n"
  gps += fmt.Sprintf ( "set title \"Latency Test: %s receivers, %d senders\\ntotal: %.2f million messages\\nZero to Two Nines\"\n", l1r.name, n_senders, millions_of_messages )
  gps += "plot \"./data\" with linespoints\n"


  // Do the plotting clause for the second graph.
  gps += fmt.Sprintf ( "set output \"%s_receivers_2.jpg\"\n", l1r.name )
  gps += fmt.Sprintf ( "set xrange [ TWO_NINE : %.7f ]\n", l1r.max_flight_time )
  gps += fmt.Sprintf ( "set title \"Latency Test: %s receivers, %d senders\\ntotal: %.2f million messages\\nTwo Nines to Max\"\n", l1r.name, n_senders, millions_of_messages )
  gps += "plot \"./data\" with linespoints\n"

  // TODO -- get this name right.
  gnuplot_script_name := "./gplot"
  file, err := os.Create ( gnuplot_script_name )
  if err != nil {
    fp ( os.Stdout, "Can't create gplot script file.\n" )
    os.Exit ( 1 )
  }
  defer file.Close()
  fmt.Fprintf ( file, "%s", gps )

  // And now write the data file.
  data_file, err := os.Create ( "./data" )
  if err != nil {
    fp ( os.Stdout, "Can't create data file.\n" )
    os.Exit ( 1 )
  }
  defer data_file.Close()

  for i := 0; i < len(l1r.bins); i ++ {
    fmt.Fprintf ( data_file, "%.7f %d\n", float64(i + 1) * l1r.bin_size, l1r.bins [ i ] )
  }

  // And execute gnuplot.
  args_list := [] string { gnuplot_script_name }
  command := exec.Command ( "/usr/bin/gnuplot", args_list ... )
  if command == nil {
    fp ( os.Stdout, "Can't make gnuplot command.\n" )
    os.Exit ( 1 )
  }

  err = command.Run ( )
  if err != nil {
    fp ( os.Stdout, "Error running gnuplot: |%s|\n", err.Error() )
    os.Exit ( 1 )
  }

  fp ( os.Stdout, "done.\n" )
}





func 
main ( ) {

  if len ( os.Args ) < 2 {
    fp ( os.Stdout, "I need dir path on command line.\n" )
  }

  dir := os.Args [ 1 ]

  // These are the subdirs that I expect to find under the main data
  // directory that was passed in as arg 1.  
  // Each of them should contain a bunch of files whose names start with 
  // "flight_times_" and that contain high-resolution (like 7 digits after 
  // the decimal point) floating point  numbers showing the flight time 
  // of each message, in seconds.
  tests := [] string { "30", "50", "70", "100", "150", "170", "200" }

  for _, test := range tests {
    fp ( os.Stdout, "   processing test %s\n", test )
    l1r := new_latency_1_result ( test )
    l1r.process_latency_1_data ( dir )
    l1r.run_gnuplot ( )
  }
}





