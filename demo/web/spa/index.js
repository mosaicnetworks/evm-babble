$(function () {
    
        // Globals variables
    
        var config = {};

        // These are called on page load
    
        // Get config
        $.getJSON( "config.json", function( data ) {
    
            // Write the data into our global variable.
            config = data;
    
            // Manually trigger a hashchange to start the app.
            $(window).trigger('hashchange');
        });
    
    
        // An event handler with calls the render function on every hashchange.
        // The render function will show the appropriate content of out page.
        $(window).on('hashchange', function(){
            render(decodeURI(window.location.hash));
        });
    
    
        // Navigation
    
        function render(url) {
    
           renderHeader();
           getEthAccounts();
           getBabbleStats();
    
        }    
    
        function renderHeader(){
            var headerTemplateScript = $("#header").html();
            var headerTemplate = Handlebars.compile(headerTemplateScript);
            $(document.body).append (headerTemplate(config));
        }

        function getEthAccounts(){
            $.ajax({
                  'url' : 'http://' + config.evm_host + '/accounts',
                  'type' : 'GET',
                  'converters' : {"* text": window.String, "text html": true, "text json": JSONbig.parse, "text xml": jQuery.parseXML},
                  'success' : function(data) {
                    renderEthAccounts(data.Accounts)
                  },
                  'error' : function(err) {
                      alert(JSON.stringify(err))
                  }
                });
        }

        function renderEthAccounts(accounts) {
            Handlebars.registerHelper('Balance', function(account) {
                return JSONbig.stringify(this.Balance)
            })

            var list = $('.accounts-list');            
            var accountsTemplateScript = $("#accounts-template").html();
            var accountsTemplate = Handlebars.compile(accountsTemplateScript);
            list.append (accountsTemplate(accounts));
        }

        function getBabbleStats(){
            $.ajax({
                'url' : 'http://' + config.babble_host + '/Stats',
                'type' : 'GET',
                'success' : function(data) {
                  renderBabbleStats(data)
                },
                'error' : function(err) {
                    alert(JSON.stringify(err))
                }
              });
        }

        function renderBabbleStats(stats) {
            var container = $('.babble-stats');            
            var statsTemplateScript = $("#stats-template").html();
            var statsTemplate = Handlebars.compile(statsTemplateScript);
            container.append (statsTemplate(stats));
        }   

        // Shows the error page.
        function renderErrorPage(){
            var page = $('.error');
            page.addClass('visible');
        }
    
    });