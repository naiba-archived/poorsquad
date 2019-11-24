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

function addAccount() {
    $('.tiny.account.modal').modal({
        closable: true,
        onApprove: function () {
            let success = false
            const btn = $('.tiny.account.modal .positive.button')
            const form = $('.tiny.account.modal form')
            if (btn.hasClass('loading')) {
                return success
            }
            form.children('.message').remove()
            btn.toggleClass('loading')
            const data = $('#accountForm').serializeArray().reduce(function (obj, item) {
                obj[item.name] = item.name.endsWith('_id') ? parseInt(item.value) : item.value;
                return obj;
            }, {});
            $.post('/api/account', JSON.stringify(data)).done(function (resp) {
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

function addCompany() {
    $('.tiny.company.modal').modal({
        closable: true,
        onApprove: function () {
            let success = false
            const btn = $('.tiny.company.modal .positive.button')
            const form = $('.tiny.company.modal form')
            if (btn.hasClass('loading')) {
                return success
            }
            form.children('.message').remove()
            btn.toggleClass('loading')
            const data = $('#companyForm').serializeArray().reduce(function (obj, item) {
                obj[item.name] = item.value;
                return obj;
            }, {});
            $.post('/api/company', JSON.stringify(data)).done(function (resp) {
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
