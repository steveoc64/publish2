!function(e){"function"==typeof define&&define.amd?define(["jquery"],e):e("object"==typeof exports?require("jquery"):jQuery)}(function(e){"use strict";function t(i,n){this.$element=e(i),this.options=e.extend({},t.DEFAULTS,e.isPlainObject(n)&&n),this.init()}var i=e(document),n="qor.publish2",r="enable."+n,s="disable."+n,o="click."+n,a="change."+n,d="qor.selectone.selected qor.selectone.unselected",l="added.qor.replicator",c="ShareableVersion",h=".qor-publish2__version",u="qor-table__inner-list",p="qor-table__inner-block",f="."+u,b="."+p,v=".qor-pulish2__eventid",g=".qor-pulish2__eventid-input",m='[name="QorResource.PublishReady"]',q='[name="QorResource.ScheduledStartAt"]',y='[name="QorResource.ScheduledEndAt"]',_='[name="QorResource.VersionName"]',S='[name="QorResource.ScheduledEventID"]',E=".qor-pulish2__action",T=".qor-pulish2__action-sharedversion",A=".qor-pulish2__action-start",k=".qor-pulish2__action-end",C=".qor-pulish2__action-input",V=".qor-action__picker-button",w=".qor-table--medialibrary>tbody>tr",L="qor-table--medialibrary",j="is-showing";return t.prototype={constructor:t,init:function(){this.actionType=this.options.element,this.bind(),this.initActionTemplate()},bind:function(){i.on(o,h,this.loadPublishVersion.bind(this)).on(a,C,this.action.bind(this)).on(d,v,this.eventidChanged.bind(this)).on(l,this.replicatorAdded.bind(this))},unbind:function(){i.off(o,h,this.loadPublishVersion.bind(this)).off(a,C,this.action.bind(this)).off(d,v,this.eventidChanged.bind(this)).off(l,this.replicatorAdded.bind(this))},initActionTemplate:function(){e(E).closest(".qor-slideout").size()||e(E).prependTo(e(".mdl-layout__content .qor-page__body")).show(),t.initSharedVersion()},replicatorAdded:function(e,i){t.generateSharedVersionLabel(i)},action:function(t){var i=e(t.target),n="checkbox"==i.prop("type"),r=i.val(),s=e(this.actionType[i.data().actionType]),o=s.closest("label");s.size()&&(n?(s.prop("checked",i.is(":checked")),i.is(":checked")?o.addClass("is-checked"):o.removeClass("is-checked")):s.val(r))},eventidChanged:function(t,i){i?e(g).val(i.primaryKey):e(g).val(""),this.updateDate(i,t.target),e(g).trigger("change")},updateDate:function(t,i){var n=e(i).closest(E),r=n.find(A),s=n.find(k),o=n.find(V),a=o.parent().find("input");t?(r.val(t.ScheduledStartAt),s.val(t.ScheduledEndAt),o.hide(),a.attr("disabled",!0)):(o.show(),a.attr("disabled",!1).closest(".is-disabled").removeClass("is-disabled")),r.trigger("change"),s.trigger("change")},loadPublishVersion:function(t){var i,n=e(t.target).parent("a"),r=n.data().versionUrl,s=n.closest("table"),o=n.closest("tr"),a=o.find("td").size(),d=s.hasClass(L),l=e('<tr class="'+u+'"><td colspan="'+a+'"></td></tr>'),c=e('<div class="'+p+'"><div style="text-align: center;"><div class="mdl-spinner mdl-js-spinner is-active"></div></div></div>');if(o.hasClass(j))return e(f).remove(),s.find("tr").removeClass(j),!1;if(e(f).remove(),s.find("tr").removeClass(j),o.addClass(j),d){var h=e(w),v=parseInt(s.width()/217),g=h.index(o)+1,m=Math.ceil(g/v);o=e(h.get(v*m-1)),o.size()||(o=h.last()),l=e('<tr class="'+u+'" style="width: '+(217*v-16)+'px"><td></td></tr>')}return o.after(l),i=e(f).find("td"),c.appendTo(i).trigger("enable"),r&&e.get(r,function(t){e(b).html(t).trigger("enable")}),!1},destroy:function(){this.unbind(),this.$element.removeData(n)}},t.generateSharedVersionLabel=function(t){var i,n=e('[name="shared-version-checkbox"]').html(),r=e('input[name$="'+c+'"]'),s={};t&&(r=t.find('input[name$="'+c+'"]')),r.each(function(){var r,a=e(this),d=a.closest(".qor-fieldset");d.hasClass(".qor-fieldset--new")||(t&&d.find(T).remove(),i=(Math.random()+1).toString(36).substring(7),s.id=[c,i].join("_"),r=e(window.Mustache.render(n,s)),r.find("input").on(o,function(){e(this).is(":checked")?a.val("true"):a.val("")}),"true"==a.val()&&r.find("input").prop("checked",!0),r.prependTo(d).trigger("enable"),a.closest(".qor-field").hide())})},t.initSharedVersion=function(){e(E).size()&&t.generateSharedVersionLabel()},e.fn.qorSliderAfterShow.initSharedVersion=t.initSharedVersion,e.fn.qorSliderAfterShow.initPublishForm=function(){var i=e(E),n=i.find("[data-action-type]"),r=t.ELEMENT;i.size()&&n.size()&&(n.each(function(){var t=e(this);e(r[t.data().actionType]).closest(".qor-form-section").hide()}),e(C).trigger(a))},t.DEFAULTS={},t.ELEMENT={scheduledstart:q,scheduledend:y,publishready:m,versionname:_,eventid:S},t.plugin=function(i){return this.each(function(){var r,s=e(this),o=s.data(n);if(!o){if(/destroy/.test(i))return;s.data(n,o=new t(this,i))}"string"==typeof i&&e.isFunction(r=o[i])&&r.apply(o)})},e(function(){var i=".qor-theme-publish2",n={};n.element=t.ELEMENT,e(document).on(s,function(n){t.plugin.call(e(i,n.target),"destroy")}).on(r,function(r){t.plugin.call(e(i,r.target),n)}).triggerHandler(r)}),t});