import ace from "ace-builds/src-noconflict/ace";

ace.define("ace/mode/ldif", function(require, exports) {
  const TextMode = require("ace/mode/text").Mode;
  const Tokenizer = require("ace/tokenizer").Tokenizer;
  const TextHighlightRules = require("ace/mode/text_highlight_rules").TextHighlightRules;

  // Proper highlight rules class
  class LdifHighlightRules extends TextHighlightRules {
    constructor() {
      super();
      this.$rules = {
        start: [
          { token: "dn", regex: "^dn:.*$" },
          { token: "keyword", regex: "^[a-zA-Z]+:" },
          { token: "comment", regex: "^#.*$" },
          { token: "text", regex: ".+" }
        ]
      };
    }
  }

  exports.Mode = class LdifMode extends TextMode {
    constructor() {
      super();
      this.HighlightRules = LdifHighlightRules;
      this.$id = "ace/mode/ldif";
    }
  };
});
