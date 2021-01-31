function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError("Cannot call a class as a function"); } }

function _defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ("value" in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } }

function _createClass(Constructor, protoProps, staticProps) { if (protoProps) _defineProperties(Constructor.prototype, protoProps); if (staticProps) _defineProperties(Constructor, staticProps); return Constructor; }

(function (cb) {
  var _this3 = this;

  // TODO : there are more cases and gotcha's.
  var wrongThis = false;

  function a1() {
    var _this = this;

    if (this !== window) {
      wrongThis = true;
    }

    var a2 = function a2() {
      if (_this !== window) {
        wrongThis = true;
      }
    };

    a2();
  }

  function b1() {
    if (this !== window) {
      wrongThis = true;
    }

    var b2 = function b2() {
      if (this !== window) {
        wrongThis = true;
      }
    };

    b2();
  }

  function c1() {
    "use strict";

    var _this2 = this;

    if (typeof this !== "undefined") {
      wrongThis = true;
    }

    var c2 = function c2() {
      if (typeof _this2 !== "undefined") {
        wrongThis = true;
      }
    };

    c2();
  }

  function d1() {
    "use strict";

    if (typeof this !== "undefined") {
      wrongThis = true;
    }

    var d2 = function d2() {
      if (typeof this !== "undefined") {
        wrongThis = true;
      }
    };

    d2();
  }

  var F1 = /*#__PURE__*/function () {
    function F1() {
      _classCallCheck(this, F1);
    }

    _createClass(F1, [{
      key: "f2",
      value: function f2() {
        if (!(this instanceof F1)) {
          wrongThis = true;
        }
      }
    }, {
      key: "f3",
      value: function f3(cb) {
        cb();
      }
    }, {
      key: "f4",
      value: function f4(cb) {
        cb.bind(this)();
      }
    }]);

    return F1;
  }();

  a1();
  b1();
  c1();
  d1();
  var f1 = new F1();
  f1.f2();
  f1.f3(function () {
    if (this !== window) {
      wrongThis = true;
    }
  });
  f1.f3(function () {
    if (_this3 !== window) {
      wrongThis = true;
    }
  });
  f1.f4(function () {
    if (this !== f1) {
      console.log(this);
      wrongThis = true;
    }
  });
  f1.f4(function () {
    if (_this3 !== window) {
      console.log(_this3);
      wrongThis = true;
    }
  });
  cb(!wrongThis);
})(callback);
