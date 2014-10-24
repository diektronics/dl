angular.module('downApp', [])
  .controller('DownCtrl', ['$scope', '$http', function($scope, $http) {
    $scope.downs = [];
    $scope.working = false;

    var logError = function(data, status) {
      console.log('code '+status+': '+data);
      $scope.working = false;
    };

    var refresh = function() {
      return $http.get('/down/').
        success(function(data) { $scope.downs = data.Downs; }).
        error(logError);
    };

    $scope.addTodo = function() {
      $scope.working = true;
      $http.post('/task/', {Title: $scope.todoText}).
        error(logError).
        success(function() {
          refresh().then(function() {
            $scope.working = false;
            $scope.todoText = '';
          })
        });
    };

    $scope.expand = function(down) {
      data = {ID: down.ID, Name: down.Name, Status: !down.Status}
      // $http.put('/task/'+task.ID, data).
      //   error(logError).
      //   success(function() { task.Done = !task.Done });
    };

    refresh().then(function() { $scope.working = false; });
  
}]);