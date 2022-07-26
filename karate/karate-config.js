function fn() {
    var config = { };
    karate.configure('logPrettyResponse', true);
    karate.configure('report', { showLog: false, showAllSteps: false } );
    return config;
  }