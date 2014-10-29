angular.module('downApp', [])
  .controller('DownCtrl', ['$scope', '$http', function($scope, $http) {
    $scope.downs = [];
    $scope.hooks = [];
    $scope.downHooks = {};
    $scope.working = false;
    $scope.visible = {};

    var logError = function(data, status) {
      console.log('code '+status+': '+data);
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
          })
        });
    };

    $scope.toggleVisibility = function(down) {
      $scope.visible[down.ID] = !$scope.visible[down.ID];
    };

    $scope.moreOrLess = function(down) {
      return $scope.visible[down.ID] ? '-' : '+';
    }

    getHooks();
    refresh().then(function() { $scope.working = false; });
  
}]);