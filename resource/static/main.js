function showConfirm(title, content, callFn, extData) {
    const modal = $('.mini.confirm.modal')
    modal.children('.header').text(title)
    modal.children('.content').text(content)
    modal.modal({
        closable: true,
        onApprove: function () {
            callFn(extData)
            return true
        }
    }).modal('show')
}

function showFormModal(modelSelector, formID, URL) {
    $(modelSelector).modal({
        closable: true,
        onApprove: function () {
            let success = false
            const btn = $(modelSelector + ' .positive.button')
            const form = $(modelSelector + ' form')
            if (btn.hasClass('loading')) {
                return success
            }
            form.children('.message').remove()
            btn.toggleClass('loading')
            const data = $(formID).serializeArray().reduce(function (obj, item) {
                obj[item.name] = item.name.endsWith('_id') ? parseInt(item.value) : item.value;
                return obj;
            }, {});
            $.post(URL, JSON.stringify(data)).done(function (resp) {
                if (resp.code == 200) {
                    window.location.reload()
                } else {
                    form.append(`<div class="ui negative message"><div class="header">操作失败</div><p>` + resp.message + `</p></div>`)
                }
            }).fail(function (err) {
                form.append(`<div class="ui negative message"><div class="header">网络错误</div><p>` + err.responseText + `</p></div>`)
            }).always(function () {
                btn.toggleClass('loading')
            });
            return success
        }
    }).modal('show')
}

function addTeam() {
    showFormModal('.tiny.team.modal', '#teamForm', '/api/team');
}

function addAccount() {
    showFormModal('.tiny.account.modal', '#accountForm', '/api/account');
}

function addCompany() {
    showFormModal('.tiny.company.modal', '#companyForm', '/api/company');
}

function logout(id) {
    $.post('/api/logout', JSON.stringify({ id: id })).done(function (resp) {
        if (resp.code == 200) {
            alert('注销成功')
            window.location.href = '/login'
        } else {
            alert('注销失败 ' + resp.code + '：' + resp.message)
        }
    }).fail(function (err) {
        alert('网络错误：' + err.responseText)
    })
}
