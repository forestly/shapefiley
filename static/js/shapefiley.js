angular.module('shapefileyApp', ['ngRoute', 'angularFileUpload'])

    .config(function($routeProvider) {
        $routeProvider
            .when('/', {
                controller:'UploadCtrl',
                templateUrl:'../upload.html'
            })
            .when('/shapefiles/:shapefileId', {
                controller:'ShowShapefileCtrl',
                templateUrl:'../show.html'
            })
            .otherwise({
                redirectTo:'/'
            });
    })
    .controller('UploadCtrl', function($scope, $upload, $http) {
        $scope.uploadProgress = 0;
        $scope.shapeFiles = [];

        $scope.onFileSelect = function($files) {
            //$files: an array of files selected, each file has name, size, and type.
            for (var i = 0; i < $files.length; i++) {
                var file = $files[i];
                $scope.upload = $upload.upload({
                    url: 'upload', //upload.php script, node.js route, or servlet url
                    method: 'POST', // or 'PUT',
                    //headers: {'header-key': 'header-value'},
                    //withCredentials: true,
                    // data: {myObj: $scope.myModelObj},
                    file: file, // or list of files ($files) for html5 only
                    //fileName: 'doc.jpg' or ['1.jpg', '2.jpg', ...] // to modify the name of the file(s)
                    // customize file formData name ('Content-Disposition'), server side file variable name.
                    //fileFormDataName: myFile, //or a list of names for multiple files (html5). Default is 'file'
                    // customize how data is added to formData. See #40#issuecomment-28612000 for sample code
                    //formDataAppender: function(formData, key, val){}
                }).progress(function(evt) {
                    $scope.uploadProgress = parseInt(100.0 * evt.loaded / evt.total);
                    console.log('percent: ' + $scope.uploadProgress);
                }).success(function(data, status, headers, config) {
                    // file is uploaded successfully
                    console.log(data);
                    $scope.shapeFiles.push(data)
                });
                //.error(...)
                //.then(success, error, progress);
                // access or attach event listeners to the underlying XMLHttpRequest.
                //.xhr(function(xhr){xhr.upload.addEventListener(...)})
            }
            /* alternative way of uploading, send the file binary with the file's content-type.
               Could be used to upload files to CouchDB, imgur, etc... html5 FileReader is needed.
               It could also be used to monitor the progress of a normal http post/put request with large data*/
            // $scope.upload = $upload.http({...})  see 88#issuecomment-31366487 for sample code.
        };
    })
    .controller('ShowShapefileCtrl', function($scope, $http, $routeParams) {
        // (function tick() {
        //     $scope.data = Data.query(function(){
        //         $timeout(tick, 1000);
        //     });
        // })();

        $scope.shapefile = {}

        var mapOptions = {
            zoom: 5,
            center: new google.maps.LatLng(37.09024, -95.712891),
            mapTypeId: google.maps.MapTypeId.TERRAIN
        };


        var plotPolygon = function(geom) {
            console.log("plotpolygon");
            for(var i = 0; i < geom.coordinates.length; i++) {
                for(var j = 0; j < geom.coordinates[i].length; j++) {
                    var coords = [];

                    for(var k = 0; k < geom.coordinates[i][j].length; k++) {
                        coords.push(
                            new google.maps.LatLng(geom.coordinates[i][j][k][1],
                                                   geom.coordinates[i][j][k][0]));

                    }

                    // Construct the polygon.
                    polygon = new google.maps.Polygon({
                        paths: coords,
                        strokeColor: '#FF0000',
                        strokeOpacity: 0.8,
                        strokeWeight: 2,
                        fillColor: '#FF0000',
                        fillOpacity: 0.35
                    });

                    polygon.setMap($scope.map);

                }
            }
        }

        $scope.map = new google.maps.Map(document.getElementById('map-container'), mapOptions);

        $http.get('/shapefiles/' + $routeParams.shapefileId).
            success(function(data, status, headers, config) {
                $scope.shapefile = data;

                if($scope.shapefile.Geom != null) {
                    for(var i = 0; i < data.Geom.length; i++) {
                        console.log(i);
                        plotPolygon(JSON.parse($scope.shapefile.Geom[i]));
                    }
                }
            }).error(function(data, status, headers, config) {});
    });
