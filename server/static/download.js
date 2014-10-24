angular.module('downApp', [])
  .controller('DownCtrl', ['$scope', '$http', function($scope, $http) {
    $scope.downs = [];
    $scope.hooks = [];
    $scope.downHooks = {};
    $scope.working = false;

    var logError = function(data, status) {
      console.log('code '+status+': '+data);
      $scope.working = false;
    };

    var getHooks = function() {
      return $http.get('/hook/').
        success(function(data) { $scope.hooks = data.Hooks; }).
        error(logError);
    };

    var refresh = function() {
      return $http.get('/down/').
        success(function(data) { $scope.downs = data.Downs; }).
        error(logError);
    };

    $scope.addDownload = function() {
      $scope.working = true;
      $http.post('/down/', {Name: $scope.downName, Links: $scope.downLinks, Hooks: $scope.downHooks}).
        error(logError).
        success(function() {
          refresh().then(function() {
            $scope.working = false;
            //$scope.downName = '';
            //$scope.downLinks = '';
          })
        });
    };

    $scope.expand = function(down) {
      data = {ID: down.ID, Name: down.Name, Status: !down.Status}
      // $http.put('/task/'+task.ID, data).
      //   error(logError).
      //   success(function() { task.Done = !task.Done });
    };

    getHooks();
    refresh().then(function() { $scope.working = false; });
  
}]);