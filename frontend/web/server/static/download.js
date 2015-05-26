angular.module('downApp', [])
  .controller('DownCtrl', ['$scope', '$http', '$interval', function($scope, $http, $interval) {
    $scope.downs = [];
    $scope.hooks = [];
    $scope.downHooks = {};
    $scope.working = false;
    $scope.visible = {};
    $scope.retError = '';

    var logError = function(data, status) {
      console.log('code '+status+': '+data);
      $scope.retError = data;
      $scope.working = false;
    };

    var getHooks = function() {
      return $http.get('hook/').
        success(function(data) { $scope.hooks = data.Hooks; }).
        error(logError);
    };

    var refresh = function() {
      return $http.get('down/').
        success(function(data) { $scope.downs = data.Downs; }).
        error(logError);
    };

    $scope.addDownload = function() {
      $scope.working = true;
      $http.post('down/', {Name: $scope.downName, Links: $scope.downLinks, Hooks: $scope.downHooks}).
        error(logError).
        success(function() {
          refresh().then(function() {
            $scope.working = false;
            $scope.downName = '';
            $scope.downLinks = '';
            $scope.downHooks = {};
            $scope.retError = ''
          })
        });
    };

    $scope.delDownload = function(down) {
      $scope.working = true;
      $http.delete('down/' + down.id).
        error(logError).
        success(function() {
          refresh().then(function() {
            $scope.working = false;
          })
        });
    };

    $scope.toggleVisibility = function(down) {
      $scope.visible[down.id] = !$scope.visible[down.id];
    };

    $scope.moreOrLess = function(down) {
      return $scope.visible[down.id] ? '-' : '+';
    }

    $scope.globalProgress = function(down){
      var total = 0.0;
      for (var i = 0; i < down.links.length; i++) {
        var percent = down.links[i].percent || 0.0;
        total += percent
      }
      return total / down.links.length; 
    }

    getHooks();
    refresh().then(function() { $scope.working = false; });
  
    var autoRefresh = $interval(function() {
      refresh();
    }, 5000);
}]);
