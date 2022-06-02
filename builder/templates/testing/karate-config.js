function fn() {
    karate.configure('logPrettyResponse', true);
    var host = karate.properties['test.server'] || 'http://127.0.0.1:9191';    
    var config = { demoBaseUrl: host };
    return config;
  }